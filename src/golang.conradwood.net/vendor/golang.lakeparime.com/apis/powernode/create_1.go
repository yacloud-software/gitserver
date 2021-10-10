// client create: PowerNodeClient
/* geninfo:
   filename  : golang.lakeparime.com/apis/powernode/powernode.proto
   gopackage : golang.lakeparime.com/apis/powernode
   importname: ai_0
   varname   : client_PowerNodeClient_0
   clientname: PowerNodeClient
   servername: PowerNodeServer
   gscvname  : powernode.PowerNode
   lockname  : lock_PowerNodeClient_0
   activename: active_PowerNodeClient_0
*/

package powernode

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_PowerNodeClient_0 sync.Mutex
  client_PowerNodeClient_0 PowerNodeClient
)

func GetPowerNodeClient() PowerNodeClient { 
    if client_PowerNodeClient_0 != nil {
        return client_PowerNodeClient_0
    }

    lock_PowerNodeClient_0.Lock() 
    if client_PowerNodeClient_0 != nil {
       lock_PowerNodeClient_0.Unlock()
       return client_PowerNodeClient_0
    }

    client_PowerNodeClient_0 = NewPowerNodeClient(client.Connect("powernode.PowerNode"))
    lock_PowerNodeClient_0.Unlock()
    return client_PowerNodeClient_0
}

