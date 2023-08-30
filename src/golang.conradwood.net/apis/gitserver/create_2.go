// client create: GITCredentialsClient
/*
  Created by /home/cnw/devel/go/yatools/src/golang.yacloud.eu/yatools/protoc-gen-cnw/protoc-gen-cnw.go
*/

/* geninfo:
   filename  : protos/golang.conradwood.net/apis/gitserver/gitserver.proto
   gopackage : golang.conradwood.net/apis/gitserver
   importname: ai_1
   clientfunc: GetGITCredentials
   serverfunc: NewGITCredentials
   lookupfunc: GITCredentialsLookupID
   varname   : client_GITCredentialsClient_1
   clientname: GITCredentialsClient
   servername: GITCredentialsServer
   gsvcname  : gitserver.GITCredentials
   lockname  : lock_GITCredentialsClient_1
   activename: active_GITCredentialsClient_1
*/

package gitserver

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_GITCredentialsClient_1 sync.Mutex
  client_GITCredentialsClient_1 GITCredentialsClient
)

func GetGITCredentialsClient() GITCredentialsClient { 
    if client_GITCredentialsClient_1 != nil {
        return client_GITCredentialsClient_1
    }

    lock_GITCredentialsClient_1.Lock() 
    if client_GITCredentialsClient_1 != nil {
       lock_GITCredentialsClient_1.Unlock()
       return client_GITCredentialsClient_1
    }

    client_GITCredentialsClient_1 = NewGITCredentialsClient(client.Connect(GITCredentialsLookupID()))
    lock_GITCredentialsClient_1.Unlock()
    return client_GITCredentialsClient_1
}

func GITCredentialsLookupID() string { return "gitserver.GITCredentials" } // returns the ID suitable for lookup in the registry. treat as opaque, subject to change.

func init() {
   client.RegisterDependency("gitserver.GITCredentials")
}
