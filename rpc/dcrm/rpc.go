/*
 *  Copyright (C) 2018-2019  Fusion Foundation Ltd. All rights reserved.
 *  Copyright (C) 2018-2019  caihaijun@fusion.org
 *
 *  This library is free software; you can redistribute it and/or
 *  modify it under the Apache License, Version 2.0.
 *
 *  This library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  
 *
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package dcrm 

import (
	"fmt"
	"time"
	"os"
	"strconv"
	"os/signal"
	"net"
	"strings"
	"github.com/fsn-dev/dcrm-sdk/rpc"
	"github.com/fsn-dev/dcrm-sdk/crypto/dcrm"
)

func listenSignal(exit chan int) {
	sig := make(chan os.Signal)
	signal.Notify(sig)

	for {
		<-sig
		exit <- 1
	}
}


type Service struct {}

// this will be called by dcrm_genPubkey
func (this *Service) GenPubkey() map[string]interface{} {   //函数名首字母必须大写
    fmt.Println("==============dcrm_genPubkey==================")
    keytype := "ECDSA"  //tmp
    if (!strings.EqualFold(keytype,"ECDSA") && !strings.EqualFold(keytype,"ED25519")) || keytype == "" {
	return map[string]interface{}{
		"error": "keytype not supported.",
	}
    }

    var err error
    for i:= 0;i<3;i++ {
	pubkey,err := dcrm.SendReqToGroup(keytype,"rpc_req_dcrmaddr")
	if err == nil && pubkey != "" {
	    return map[string]interface{}{
		    "pubkey": pubkey,
	    }
	}
	
	time.Sleep(time.Duration(1000000)) //1000 000 000 == 1s
    }
    
    if err != nil {
	fmt.Println("===========dcrm_genPubkey,err=%v============",err)
	return map[string]interface{}{
		"error": err.Error(),
	}
    }
    
    return map[string]interface{}{
	    "error":"gen pubkey fail.",
    }
    
}

// this will be called by dcrm_sign
func (this *Service) Sign(pubkey string,message string) map[string]interface{} {
    fmt.Println("==============dcrm_sign==================")
    keytype := "ECDSA"  //tmp
    if pubkey == "" || message == "" {
	return map[string]interface{}{
		"error": "pubkey is empty.",
	}
    }

    if (!strings.EqualFold(keytype,"ECDSA") && !strings.EqualFold(keytype,"ED25519")) || keytype == "" {
	return map[string]interface{}{
		"error": "keytype not supported.",
	}
    }

    var err error
    for i:=0;i<3;i++ {
	msg := pubkey + ":" + keytype + ":" + message
	rsv,err := dcrm.SendReqToGroup(msg,"rpc_sign")
	if err == nil && rsv != "" {
	    fmt.Println("========dcrm_sign,rsv = %s =============",rsv)
	    return map[string]interface{}{
		    "rsv": rsv,
	    }
	}
	
	time.Sleep(time.Duration(1000000)) //1000 000 000 == 1s
    }

    if err != nil {
	fmt.Println("========dcrm_sign,err = %s =============",err.Error())
	return map[string]interface{}{
		"error": err.Error(),
	}
    }
    
    return map[string]interface{}{
	    "error": "dcrm sign fail.",
    }
    
}

var (
	rpcport  int
	endpoint string = "0.0.0.0"
	server *rpc.Server
	err      error
)

func RpcInit(port int) {
	rpcport = port
	go startRpcServer()
}

// splitAndTrim splits input separated by a comma
// and trims excessive white space from the substrings.
func splitAndTrim(input string) []string {
	result := strings.Split(input, ",")
	for i, r := range result {
		result[i] = strings.TrimSpace(r)
	}
	return result
}

func startRpcServer() error {
	go func() error {
	    server = rpc.NewServer()
	    service := new(Service)
	    if err := server.RegisterName("dcrm", service); err != nil {
		    panic(err)
	    }

	    // All APIs registered, start the HTTP listener
	    var (
		    listener net.Listener
		    err      error
	    )

	    endpoint = endpoint + ":" + strconv.Itoa(rpcport)
	    if listener, err = net.Listen("tcp", endpoint); err != nil {
		    panic(err)
	    }

	    /////////
	    /*
	    var (
		    extapiURL = "n/a"
		    ipcapiURL = "n/a"
	    )
	    rpcAPI := []rpc.API{
		    {
			    Namespace: "account",
			    Public:    true,
			    Service:   api,
			    Version:   "1.0"},
	    }
	    if c.Bool(utils.RPCEnabledFlag.Name) {

		    vhosts := splitAndTrim(c.GlobalString(utils.RPCVirtualHostsFlag.Name))
		    cors := splitAndTrim(c.GlobalString(utils.RPCCORSDomainFlag.Name))

		    // start http server
		    httpEndpoint := fmt.Sprintf("%s:%d", c.String(utils.RPCListenAddrFlag.Name), c.Int(rpcPortFlag.Name))
		    listener, _, err := rpc.StartHTTPEndpoint(httpEndpoint, rpcAPI, []string{"account"}, cors, vhosts, rpc.DefaultHTTPTimeouts)
		    if err != nil {
			    utils.Fatalf("Could not start RPC api: %v", err)
		    }
		    extapiURL = fmt.Sprintf("http://%s", httpEndpoint)
		    log.Info("HTTP endpoint opened", "url", extapiURL)

		    defer func() {
			    listener.Close()
			    log.Info("HTTP endpoint closed", "url", httpEndpoint)
		    }()

	    }
	    if !c.Bool(utils.IPCDisabledFlag.Name) {
		    if c.IsSet(utils.IPCPathFlag.Name) {
			    ipcapiURL = c.String(utils.IPCPathFlag.Name)
		    } else {
			    ipcapiURL = filepath.Join(configDir, "clef.ipc")
		    }

		    listener, _, err := rpc.StartIPCEndpoint(ipcapiURL, rpcAPI)
		    if err != nil {
			    utils.Fatalf("Could not start IPC api: %v", err)
		    }
		    log.Info("IPC endpoint opened", "url", ipcapiURL)
		    defer func() {
			    listener.Close()
			    log.Info("IPC endpoint closed", "url", ipcapiURL)
		    }()

	    }
	    */
	    /////////

	    vhosts := make([]string, 0)
	    cors := splitAndTrim("*")
	    go rpc.NewHTTPServer(cors, vhosts, rpc.DefaultHTTPTimeouts,server).Serve(listener)
	    rpcstring := "\n==================== RPC Service Already Start! url = " + fmt.Sprintf("http://%s", endpoint) + " =====================\n"
	    fmt.Println(rpcstring)

	    exit := make(chan int)
	    <-exit

	    server.Stop()

	    return nil
	}()

	return nil
}

