// Copyright 2019 HuaweiCloud.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"time"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/servers"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)


var outFile = flag.String("config.write-to", "ecs_file_sd.yml", "path of file to write ECS service discovery information to")
var interval = flag.Duration("config.scrape-interval", 60*time.Second, "interval at which to scrape the Huaweicloud API for ECS service discovery information")
var times = flag.Int("config.scrape-times", 0, "how many times to scrape before exiting (0 = infinite)")

var projectName = flag.String("config.projectName", "", "The Name of the Tenant (Identity v2) or Project (Identity v3) to login with.")
var userName = flag.String("config.userName", "", "The Username to login with.")
var accessKey = flag.String("config.accessKey", "", "The access key of the HuaweiCloud to use (optional)")
var secretKey = flag.String("config.secretKey", "", "The secret key of the HuaweiCloud to use.")
var domain = flag.String("config.domain", "", "The Name of the Domain to scope to (Identity v3).")
var region = flag.String("config.region", "", "The region of the HuaweiCloud to use")
var password = flag.String("config.password", "", "The Password to login with.")
var port = flag.String("config.port", "9100", "")
var isHuaweicloudModule = flag.Bool("config.model", false, "If the config.model is set to true, the model LabelName will added MetaLabelPrefix(__meta_huaweicloud_)")

const (
	huaweicloudLabelPrefix         = model.MetaLabelPrefix + "huaweicloud_"
	huaweicloudLabelAddressPool    = huaweicloudLabelPrefix + "address_pool"
	huaweicloudLabelInstanceFlavor = huaweicloudLabelPrefix + "instance_flavor"
	huaweicloudLabelInstanceID     = huaweicloudLabelPrefix + "instance_id"
	huaweicloudLabelInstanceName   = huaweicloudLabelPrefix + "instance_name"
	huaweicloudLabelInstanceStatus = huaweicloudLabelPrefix + "instance_status"
	huaweicloudLabelPrivateIP      = huaweicloudLabelPrefix + "private_ip"
	huaweicloudLabelProjectID      = huaweicloudLabelPrefix + "project_id"
	huaweicloudLabelUserID         = huaweicloudLabelPrefix + "user_id"
)

type Config struct {
	AccessKey        string
	SecretKey        string
	DomainID         string
	DomainName       string
	EndpointType     string
	IdentityEndpoint string
	Insecure         bool
	Password         string
	Region           string
	TenantID         string
	TenantName       string
	Token            string
	Username         string
	UserID           string

	HwClient *golangsdk.ProviderClient
}

type Labels struct {
	Name  string `json:"name"`
}

type PrometheusInfo struct {
	Targets []string  `json:"targets"`
	Labels  Labels    `json:"labels"`
}

func buildClient(c *Config) error {
	err := fmt.Errorf("Must config token or aksk or username password to be authorized")

	if c.AccessKey != "" && c.SecretKey != "" {
		err = buildClientByAKSK(c)
	} else if c.Password != "" && (c.Username != "" || c.UserID != "") {
		err = buildClientByPassword(c)
	}

	if err != nil {
		return err
	}

	return nil
}


func buildClientByPassword(c *Config) error {
	var pao, dao golangsdk.AuthOptions

	pao = golangsdk.AuthOptions{
		DomainID:   c.DomainID,
		DomainName: c.DomainName,
		TenantID:   c.TenantID,
		TenantName: c.TenantName,
	}

	dao = golangsdk.AuthOptions{
		DomainID:   c.DomainID,
		DomainName: c.DomainName,
	}

	for _, ao := range []*golangsdk.AuthOptions{&pao, &dao} {
		ao.IdentityEndpoint = c.IdentityEndpoint
		ao.Password = c.Password
		ao.Username = c.Username
		ao.UserID = c.UserID
	}

	return genClients(c, pao, dao)
}

func buildClientByAKSK(c *Config) error {
	var pao, dao golangsdk.AKSKAuthOptions

	pao = golangsdk.AKSKAuthOptions{
		ProjectName: c.TenantName,
		ProjectId:   c.TenantID,
	}

	dao = golangsdk.AKSKAuthOptions{
		DomainID: c.DomainID,
		Domain:   c.DomainName,
	}

	for _, ao := range []*golangsdk.AKSKAuthOptions{&pao, &dao} {
		ao.IdentityEndpoint = c.IdentityEndpoint
		ao.AccessKey = c.AccessKey
		ao.SecretKey = c.SecretKey
	}
	return genClients(c, pao, dao)
}


func genClients(c *Config, pao, dao golangsdk.AuthOptionsProvider) error {
	client, err := genClient(c, pao)
	if err != nil {
		return err
	}
	c.HwClient = client
	return err
}

func genClient(c *Config, ao golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
	client, err := openstack.NewClient(ao.GetIdentityEndpoint())
	if err != nil {
		return nil, err
	}

	client.HTTPClient = http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if client.AKSKAuthOptions.AccessKey != "" {
				golangsdk.ReSign(req, golangsdk.SignOptions{
					AccessKey: client.AKSKAuthOptions.AccessKey,
					SecretKey: client.AKSKAuthOptions.SecretKey,
				})
			}
			return nil
		},
	}

	err = openstack.Authenticate(client, ao)
	if err != nil {
		return nil, err
	}

	return client, nil
}


func getModelLabelsTags(allServers []servers.Server)([]*targetgroup.Group, error)  {
	tg := &targetgroup.Group{
		Source: fmt.Sprintf("OS_" + *region),
	}

	for _, server := range allServers {
		labels := model.LabelSet{
			model.LabelName(huaweicloudLabelInstanceID):     model.LabelValue(server.ID),
			model.LabelName(huaweicloudLabelInstanceStatus): model.LabelValue(server.Status),
			model.LabelName(huaweicloudLabelInstanceName):   model.LabelValue(server.Name),
			model.LabelName(huaweicloudLabelProjectID):      model.LabelValue(server.TenantID),
			model.LabelName(huaweicloudLabelUserID):         model.LabelValue(server.UserID),
		}

		id, ok := server.Flavor["id"].(string)
		if !ok {
			fmt.Println("msg", "Invalid type for flavor id, expected string")
			continue
		}
		labels[huaweicloudLabelInstanceFlavor] = model.LabelValue(id)


		for _, address := range server.Addresses {
			md, ok := address.([]interface{})
			if !ok {
				fmt.Println("msg", "Invalid type for address, expected array")
				continue
			}
			if len(md) == 0 {
				fmt.Println("msg", "Got no IP address", "instance", server.ID)
				continue
			}
			for pool, address := range md {
				md1, ok := address.(map[string]interface{})
				if !ok {
					fmt.Println("msg", "Invalid type for address, expected dict")
					continue
				}
				addr, ok := md1["addr"].(string)
				if !ok {
					fmt.Println("msg", "Invalid type for address, expected string")
					continue
				}

				lbls := make(model.LabelSet, len(labels))
				for k, v := range labels {
					lbls[k] = v
				}
				lbls[huaweicloudLabelAddressPool] = model.LabelValue(pool)
				lbls[huaweicloudLabelPrivateIP] = model.LabelValue(addr)

				addr = net.JoinHostPort(addr, fmt.Sprintf("%s", *port))
				lbls[model.AddressLabel] = model.LabelValue(addr)

				tg.Targets = append(tg.Targets, lbls)
			}
		}
	}

	return []*targetgroup.Group{tg}, nil
}

func getPi(pis []*PrometheusInfo, serverName string) (*PrometheusInfo, bool)  {
	for _, pi := range pis {
		if (pi.Labels.Name == serverName){
			return pi, true
		}
	}

	var pi PrometheusInfo
	pi.Labels = Labels{
		Name: serverName,
	}

	return &pi, false
}


func getSimpleTags(allServers []servers.Server)([]*PrometheusInfo, error) {
	pi := PrometheusInfo{
		Labels: Labels{Name: allServers[0].Name},
	}
	pis := []*PrometheusInfo{&pi}

	for _, server := range allServers {
		pi, isSame := getPi(pis, server.Name)

		for _, address := range server.Addresses {
			md, ok := address.([]interface{})
			if !ok {
				fmt.Println("msg", "Invalid type for address, expected array")
				continue
			}

			if len(md) == 0 {
				fmt.Println("msg", "Got no IP address", "instance", server.ID)
				continue
			}
			for _, address := range md {
				md1, ok := address.(map[string]interface{})
				if !ok {
					fmt.Println("msg", "Invalid type for address, expected dict")
					continue
				}
				addr, ok := md1["addr"].(string)
				if !ok {
					fmt.Println("msg", "Invalid type for address, expected string")
					continue
				}

				addr = net.JoinHostPort(addr, fmt.Sprintf("%s", *port))
				pi.Targets = append(pi.Targets, addr)
			}

			if (isSame == false) {
				pis = append(pis, pi)
			}
		}
	}

	return pis, nil
}

//List servers query server list
func ServersList(client *golangsdk.ServiceClient)([]servers.Server, error) {
	// Query all servers list information
	allPages, err := servers.List(client, servers.ListOpts{}).AllPages()
	if err != nil {
		fmt.Println("allPagesErr:", err)
		return nil, err
	}

	// Transform servers structure
	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		fmt.Println("allServersErr:", err)
		return nil, err
	}
	fmt.Println("Got servers list success")

	return allServers, nil
}

func checkConfigOptions() error  {
	if (*region == ""){
		return fmt.Errorf("The config.region should be set.")
	}

	if (*projectName == ""){
		return fmt.Errorf("The config.projectName should be set.")
	}

	if (*domain == ""){
		return fmt.Errorf("The config.domains should be set.")
	}

	if (*userName == ""){
		return fmt.Errorf("The config.userName should be set.")
	}

	return nil
}

func initClient()(*golangsdk.ServiceClient, error)  {
	configOptions := Config{
		IdentityEndpoint: "https://iam.cn-north-1.myhwclouds.com/v3",
		TenantName:      *projectName,
		AccessKey:       *accessKey,
		SecretKey:       *secretKey,
		DomainName:      *domain,
		Username:        *userName,
		Region:          *region,
		Password:        *password,
		Insecure:        true,
	}

	err := buildClient(&configOptions)
	if err != nil {
		fmt.Println("Failed to build client: ", err)
		return nil, err
	}

	//Init service client
	client, clientErr := openstack.NewComputeV2(configOptions.HwClient, golangsdk.EndpointOpts{
		Region: "cn-north-1",
	})
	if clientErr != nil {
		fmt.Println("Failed to get the NewComputeV2 client: ", clientErr)
		return nil, clientErr
	}

	return client, err
}

func main() {
	flag.Parse()

	err := checkConfigOptions()
	if err != nil {
		fmt.Println("Failed to validate config options: %s", err)
		return
	}

	work := func() {
		client, err := initClient()
		if err != nil {
			fmt.Println("Init client fail:", err)
			return
		}

		allServers, err := ServersList(client)
		if err != nil {
			fmt.Println("ServersList fail:", err)
			return
		}

		if len(allServers) == 0 {
			fmt.Println("The serversList does not find")
			return
		}

		var m []byte
		if (*isHuaweicloudModule == true) {
			tgs, err := getModelLabelsTags(allServers)
			if err != nil{
				fmt.Errorf("Can not get all model labels from huaweicloud instances: %s.", err)
				return
			}
			m , err = json.MarshalIndent(tgs, "", " ")
			if err != nil {
				return
			}
		} else {
			pis, _ := getSimpleTags(allServers)
			m, err = json.MarshalIndent(pis, "", " ")
			if err != nil {
				return
			}
		}

		log.Printf("Writing discovered exporters to %s", *outFile)
		err = ioutil.WriteFile(*outFile, m, 0644)
		if err != nil {
			return
		}
	}

	s := time.NewTimer(1 * time.Millisecond)
	t := time.NewTicker(*interval)
	n := *times

	for {
		select {
		case <-s.C:
		case <-t.C:
		}
		work()
		n = n - 1
		if *times > 0 && n == 0 {
			break
		}
	}
}
