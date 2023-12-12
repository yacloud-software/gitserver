// client create: GIT2Client
/*
  Created by /home/cnw/devel/go/yatools/src/golang.yacloud.eu/yatools/protoc-gen-cnw/protoc-gen-cnw.go
*/

/* geninfo:
   filename  : protos/golang.conradwood.net/apis/gitserver/gitserver.proto
   gopackage : golang.conradwood.net/apis/gitserver
   importname: ai_0
   clientfunc: GetGIT2
   serverfunc: NewGIT2
   lookupfunc: GIT2LookupID
   varname   : client_GIT2Client_0
   clientname: GIT2Client
   servername: GIT2Server
   gsvcname  : gitserver.GIT2
   lockname  : lock_GIT2Client_0
   activename: active_GIT2Client_0
*/

package gitserver

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_GIT2Client_0 sync.Mutex
  client_GIT2Client_0 GIT2Client
)

func GetGIT2Client() GIT2Client { 
    if client_GIT2Client_0 != nil {
        return client_GIT2Client_0
    }

    lock_GIT2Client_0.Lock() 
    if client_GIT2Client_0 != nil {
       lock_GIT2Client_0.Unlock()
       return client_GIT2Client_0
    }

    client_GIT2Client_0 = NewGIT2Client(client.Connect(GIT2LookupID()))
    lock_GIT2Client_0.Unlock()
    return client_GIT2Client_0
}

func GIT2LookupID() string { return "gitserver.GIT2" } // returns the ID suitable for lookup in the registry. treat as opaque, subject to change.

func init() {
   client.RegisterDependency("gitserver.GIT2")
   AddService("gitserver.GIT2")
}



