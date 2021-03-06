package webui

import (
	"encoding/json"
	"github.com/j-keck/lsleases/pkg/cscom"
	"github.com/j-keck/plog"
	"net/http"
	"strings"
)

var log = plog.GlobalLogger()

type endpoint struct {
	path string
	hndl func(http.ResponseWriter, *http.Request)
}

type WebUI struct {
}

func NewWebUI() WebUI {
	self := new(WebUI)
	self.registerEndpoints()
	return *self
}

func (self *WebUI) ListenAndServe(addr string) {
	browserAddr := addr
	if strings.HasPrefix(browserAddr, ":") {
		browserAddr = "http://localhost" + browserAddr
	} else {
		browserAddr = "http://" + browserAddr
	}

	log.Infof("startup webui on address: %s - you can open the webui at: %s",
		addr, browserAddr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Warnf("unable to start webui: %v", err)
	}
}

func (self *WebUI) registerEndpoints() {
	// webui
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(index))
	})

	// version
	http.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		if version, err := cscom.AskServer(cscom.GetVersion); err == nil {
			json, _ := json.Marshal(struct {
				V string `json:"version"`
			}{version.(cscom.Version).String()})
			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
		} else {
			log.Warnf("unable to lookup server version: %v", err)
		}
	})

	// leases listing
	http.HandleFunc("/api/leases", func(w http.ResponseWriter, r *http.Request) {
		var leases cscom.Leases
		var err error
		if since := r.URL.Query().Get("since"); len(since) != 0 {
			var resp cscom.ServerResponse
			resp, err = cscom.AskServerWithPayload(
				cscom.GetLeasesSince,
				since,
			)
			leases = resp.(cscom.Leases)
		} else {
			var resp cscom.ServerResponse
			resp, err = cscom.AskServer(cscom.GetLeases)
			leases = resp.(cscom.Leases)
		}

		if err == nil {
			json, _ := json.Marshal(leases)
			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
		} else {
			log.Warnf("unable to lookup leases: %v", err)
		}
	})

	// clear leases
	http.HandleFunc("/api/clear-leases", func(w http.ResponseWriter, r *http.Request) {
		cscom.TellServer(cscom.ClearLeases)
	})
}

var index = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <title>lsleases WebUI</title>
    <style>
      .container {
          display: flex;
          justify-content: center;
      }
      .content {
          flex-direction: column;
      }
      #notification {
          color: #b51515;
          margin-top: 10px;
          margin-left: 15px;
      }
      table {
          border-collapse: collapse;
      }
      th, td {
          padding: 1em;
          border-bottom: 1px solid #ddd;
      }
      tr:hover {
          background-color: #f5f5f5;
      }

      td#created {
          text-align: right;
      }
      td#ip {
          text-align: right;
      }
      #footer {
          margin-top: 5px;
          text-align: right;
          font-size: 11px;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div id="content">
        <div id="notification"></div>
        <table id="leases">
          <thead><tr><th>Captured</th><th>IP</th><th>MAC</th><th>Hostname</th></tr></thead>
          <tbody></tbody>
        </table>
        <div id="footer">lsleases</div>
      </div>
    </div>

    <script language="javascript">
      let since = 0;

      get("/api/version", function(obj) {
          let version = " v" + obj.version;
          let node = document.createElement("span");
          node.appendChild(document.createTextNode(version));
          node.style = "font-size: 8px";
          document.getElementById("footer").appendChild(node);
      });

      window.setInterval(function() {
          get("/api/leases?since=" + since, function(leases) {
              since = new Date().getTime() * 1000000;
              updateNotification("");
              leases.sort(function(a, b) { return a.Created > b.Created });
              leases.forEach(function(item, index) {
                  updateLeasesTable(item);
              });
              }, function(xhr) {
                  updateNotification("Unable to fetch leases");
              }
          )}, 1000);


        function get(path, cbOk, cbErr) {
          let xhr;
          if (window.XMLHttpRequest) {
              xhr = new XMLHttpRequest();
          } else if (window.ActiveXObject) {
              xhr = new ActiveXObject("Microsoft.XMLHTTP");
          }
          if (!xhr) {
              updateNotification("cannot create an XMLHttp instance");
          }
          xhr.onreadystatechange = function() {
              if (xhr.readyState == XMLHttpRequest.DONE) {
                  if(xhr.statusText == "OK") {
                    cbOk(JSON.parse(xhr.response));
                  } else {
                    cbErr(xhr);
                  }
              }
          };
          xhr.open("GET", path);
          xhr.send();
      }

      function updateNotification(txt) {
          document.getElementById("notification").innerHTML = txt;
      }

      function updateLeasesTable(lease) {
          let tbody = document.getElementById("leases").getElementsByTagName("tbody")[0];

          // remove the old entry
          let old = document.getElementById(lease.Mac);
          if(old != null) {
              tbody.removeChild(old);
          }

          // create a new row
          let row = tbody.insertRow(0);
          row.id = lease.Mac;

          // convert timestamp
          let created = new Date(Date.parse(lease.Created));
          let createdCell = row.insertCell();
          createdCell.id = "created";
          createdCell.title = created.toLocaleString();
          createdCell.textContent = created.toLocaleTimeString();

          // ip as link
          let ipCell = row.insertCell();
          let ipLink = document.createElement("a");
          ipLink.href = "http://" + lease.IP;
          ipLink.target = "_blank";
          ipLink.textContent = lease.IP;
          ipCell.appendChild(ipLink);

          // text cells
          ["Mac", "Host"].forEach(function(n) {
              let cell = row.insertCell();
              cell.id = n.toLowerCase();
              cell.textContent = lease[n];
          });
      }
    </script>
  </body>
`
