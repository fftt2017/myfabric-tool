package controller

import (
	"encoding/json"
	"log"
	"myfabric-tool/chain/client"
	"myfabric-tool/config"
	"net/http"
)

func OrderPeerConfigJson(w http.ResponseWriter, r *http.Request) {
	var orders []map[string]interface{}
	var peers []map[string]interface{}
	var users []map[string]interface{}

	orgsKeys := config.ListOrgsKey()

	for i := 0; i < len(orgsKeys); i++ {
		orgName := orgsKeys[i]
		org := config.GetOrg(orgName)
		nodeKeys := org.ListNodeKeys()
		for j := 0; j < len(nodeKeys); j++ {
			node := org.GetNode(nodeKeys[j])
			if node.Type == "orderer" {
				order := map[string]interface{}{
					"name": nodeKeys[j],
					"org":  orgsKeys[i],
				}
				orders = append(orders, order)
			} else if node.Type == "peer" {
				peer := map[string]interface{}{
					"name": nodeKeys[j],
					"org":  orgsKeys[i],
				}
				peers = append(peers, peer)
			}
		}

		userKeys := org.ListUserKeys()
		for j := 0; j < len(userKeys); j++ {
			users = append(users, map[string]interface{}{
				"name": userKeys[j],
				"org":  orgsKeys[i],
			})
		}
	}
	savedParam,_ :=config.GetSwitchParam()
	result := map[string]interface{}{
		"orders": orders,
		"peers":  peers,
		"users":  users,
		"saved": savedParam,
	}
	if savedParam != nil {
		client.RefreshDefaultClient(savedParam["orderOrg"], savedParam["ordererNode"], savedParam["peerOrg"], savedParam["peerNode"], savedParam["userOrg"], savedParam["user"])
	}
	b, _ := json.Marshal(result)
	log.Println(string(b))
	w.Write(b)
}

func SwitchNode(w http.ResponseWriter, r *http.Request) {
	ordererOrg := r.FormValue("ordererOrg")
	ordererNode := r.FormValue("ordererNode")
	peerOrg := r.FormValue("peerOrg")
	peerNode := r.FormValue("peerNode")
	userOrg := r.FormValue("userOrg")
	user := r.FormValue("user")
	err :=client.RefreshDefaultClient(ordererOrg,ordererNode,peerOrg,peerNode,userOrg,user)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	err = config.SaveSwitchParam(ordererOrg, ordererNode,peerOrg,peerNode,userOrg,user)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("success"))
}
