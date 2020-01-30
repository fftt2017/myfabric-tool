package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

const (
	TYPE_ORDERER = "orderer"
	TYPE_PEER    = "peer"
)

type User struct {
	MspConfigPath     string
	TLSClientKeyFile  string `mapstructure:"tls-client-key-file"`
	TLSClientCertFile string `mapstructure:"tls-client-cert-file"`
}

type Node struct {
	Type                  string
	Address               string
	TLSEnabled            bool   `mapstructure:"tls-enabled"`
	TLSServerHostOverride string `mapstructure:"tls-serverhostoverride"`
	TLSClientAuthRequired bool   `mapstructure:"tls-clientAuthRequired"`
}

type Org struct {
	LocalMSPId      string
	TLSRootCertFile string           `mapstructure:"tls-root-cert-file"`
	Users           map[string]*User `mapstructure:"_"`
	Nodes           map[string]*Node `mapstructure:"_"`
}

func (o *Org) GetUser(key string) *User {
	return o.Users[key]
}

func (o *Org) GetNode(key string) *Node {
	return o.Nodes[key]
}

func (o *Org) ListUserKeys() []string {
	keys := make([]string, 0, len(o.Users))
	for k := range o.Users {
		keys = append(keys, k)
	}
	return keys
}

func (o *Org) ListNodeKeys() []string {
	keys := make([]string, 0, len(o.Nodes))
	for k := range o.Nodes {
		keys = append(keys, k)
	}
	return keys
}

var Orgs map[string]*Org

func ListOrgsKey() []string {
	keys := make([]string, 0, len(Orgs))
	for k := range Orgs {
		keys = append(keys, k)
	}
	return keys
}

func GetOrg(key string) *Org {
	return Orgs[key]
}

func LoadConfig(file string) error {
	fp, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	yf, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}

	m := make(map[interface{}]interface{})

	if err = yaml.Unmarshal(yf, &m); err != nil {
		return err
	}

	orgs := m["orgs"].([]interface{})
	if len(orgs) == 0 {
		return errors.New("no any org")
	}
	Orgs = make(map[string]*Org, len(orgs))
	for i := range orgs {
		orgm := m[orgs[i]].(map[interface{}]interface{})
		org := new(Org)
		if err = mapstructure.Decode(orgm, org); err != nil {
			return err
		}

		if orgm["users"] != nil {
			users := orgm["users"].([]interface{})
			if len(users) == 0 {
				org.Users = make(map[string]*User, 0)
			} else {
				org.Users = make(map[string]*User, len(users))
				for i := range users {
					userm := orgm[users[i]].(map[interface{}]interface{})
					user := new(User)
					if err = mapstructure.Decode(userm, user); err != nil {
						return err
					}
					org.Users[users[i].(string)] = user
				}
			}
		}

		if orgm["nodes"] != nil {
			nodes := orgm["nodes"].([]interface{})
			if len(nodes) == 0 {
				org.Nodes = make(map[string]*Node, 0)
			} else {
				org.Nodes = make(map[string]*Node, len(nodes))
				for i := range nodes {
					nodem := orgm[nodes[i]].(map[interface{}]interface{})
					node := new(Node)
					if err = mapstructure.Decode(nodem, node); err != nil {
						return err
					}
					org.Nodes[nodes[i].(string)] = node
				}
			}
		}

		Orgs[orgs[i].(string)] = org
	}

	return nil
}

func SaveSwitchParam(ordererOrg, ordererNode, peerOrg, peerNode, userOrg, user string) (error) {
	param := map[string]interface{}{
		"orderOrg":    ordererOrg,
		"ordererNode": ordererNode,
		"peerOrg":     peerOrg,
		"peerNode":    peerNode,
		"userOrg":     userOrg,
		"user":        user,
	}
	out, err := json.Marshal(&param)
	if err !=nil {
		return err
	}

	pwd,err:=os.Getwd()
	if err!=nil {
		return err
	}
	ioutil.WriteFile(pwd+"/param.json", out, 755)
	return nil
}
func GetSwitchParam() (ret map[string]string, err error) {
	pwd,err:=os.Getwd()
	if err!=nil {
		return nil,err
	}
	fp, err := filepath.Abs(pwd+"/param.json")
	if err != nil {
		return nil,err
	}
	yf, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil,err
	}

	m := make(map[string]string)

	if err = json.Unmarshal(yf, &m); err != nil {
		return nil,err
	}
	return m,nil
}
