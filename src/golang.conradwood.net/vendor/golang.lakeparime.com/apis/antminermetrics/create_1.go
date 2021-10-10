// client create: AntMinerMetricsClient
/* geninfo:
   filename  : golang.lakeparime.com/apis/antminermetrics/antminermetrics.proto
   gopackage : golang.lakeparime.com/apis/antminermetrics
   importname: ai_0
   varname   : client_AntMinerMetricsClient_0
   clientname: AntMinerMetricsClient
   servername: AntMinerMetricsServer
   gscvname  : antminermetrics.AntMinerMetrics
   lockname  : lock_AntMinerMetricsClient_0
   activename: active_AntMinerMetricsClient_0
*/

package antminermetrics

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_AntMinerMetricsClient_0 sync.Mutex
  client_AntMinerMetricsClient_0 AntMinerMetricsClient
)

func GetAntMinerMetricsClient() AntMinerMetricsClient { 
    if client_AntMinerMetricsClient_0 != nil {
        return client_AntMinerMetricsClient_0
    }

    lock_AntMinerMetricsClient_0.Lock() 
    if client_AntMinerMetricsClient_0 != nil {
       lock_AntMinerMetricsClient_0.Unlock()
       return client_AntMinerMetricsClient_0
    }

    client_AntMinerMetricsClient_0 = NewAntMinerMetricsClient(client.Connect("antminermetrics.AntMinerMetrics"))
    lock_AntMinerMetricsClient_0.Unlock()
    return client_AntMinerMetricsClient_0
}

