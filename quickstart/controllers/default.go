package controllers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	"github.com/lubronzhan/authz-vic/quickstart/models"
	"github.com/supervised-io/tokenverifier"
)

const passwordGrantFormatString = "grant_type=password&username=%s&password=%s&scope=%s"
const tokenScope = "openid offline_access id_groups at_groups rs_admin_server"
const baseUrl = "https://%s/openidconnect/token/%s"
const certUrl = "https://%s/idm/tenant/%s/certificates/?scope=TENANT"

var (
	tenant = "vsphere.local"
	tr     = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	httpClient = &http.Client{Transport: tr}
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	c.Data["Website"] = "beego.me"
	c.Data["Email"] = "astaxie@gmail.com"
	c.Data["Name"] = "Lubron"
	c.TplName = "index.tpl"
}

func (c *MainController) Post() {
	c.Data["Name"] = "Lubron"
	c.Data["isAdmin"] = false
	c.TplName = "index.tpl"
	u := models.User{}
	if err := c.ParseForm(&u); err != nil {
		//handle error
		fmt.Println("error passing form")
	}
	// send request to LW
	c.TplName = "index.tpl"
	username := c.GetString("username")
	password := c.GetString("password")
	hostname := c.GetString("hostname")

	accessToken, err := requestToken(username, password)
	if err != nil {
		fmt.Println(err)
		return
	}

	jwtToken, err := tokenverifier.ParseTokenDetails(*accessToken)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(jwtToken.Groups)
	for _, string := range jwtToken.Groups {
		if strings.Split(string, "\\")[1] == "Administrators" {
			fmt.Println("show administrator thing")
			c.Data["isAdmin"] = true
			return
		}
	}

}

func requestToken(username string, password string) (*string, error) {
	body := fmt.Sprintf(passwordGrantFormatString, username, password, tokenScope)
	request, err := http.NewRequest("POST", buildTokenURL(hostname, tenant), strings.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(request)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	tokenResponse := &models.TokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(tokenResponse)
	if err != nil {
		fmt.Println(fmt.Sprintf("Exiting GetToken, failed decoding: %s", err.Error()))
		return nil, err
	}

	return &tokenResponse.AccessToken, nil
}

func requestCertificate(username string, password string) (*[]byte, error) {
	body := fmt.Sprintf(passwordGrantFormatString, username, password, tokenScope)
	request, err := http.NewRequest("GET", buildCertURL(hostname, tenant), strings.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(request)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	certificateByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(fmt.Sprintf("Exiting Get Certificate, failed decoding: %s", err.Error()))
		return nil, err
	}

	return &certificateByte, nil
}

func buildTokenURL(hostname string, tenant string) string {
	return fmt.Sprintf(baseUrl, hostname, tenant)
}
func buildCertURL(hostname string, tenant string) string {
	return fmt.Sprintf(certUrl, hostname, tenant)
}
