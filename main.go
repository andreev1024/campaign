/*
*	generate data
	http GET http://localhost:3000/campaign x==100 y==26 z==10000 -v

 */
package main

import (
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"campaign/res"
	"strconv"
	"time"
	"math/rand"
	"io/ioutil"
)

const (
	PORT = 3000
	CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	DATA_PERM = 0644
	DATA_NAME = "data.json"
	DATA_PATH = "./"
)



//@todo formatting + comments + refactoring
func generateCampData(w http.ResponseWriter, r *http.Request) {

	req := struct {
		x string
		y string
		z string
	}{
		x: r.URL.Query().Get("x"),
		y: r.URL.Query().Get("y"),
		z: r.URL.Query().Get("z"),
	}

	var x, y, z int
	var errs []error
	var err error
	if x, err = strconv.Atoi(req.x); err != nil {
		errs = append(errs, err)
	}
	if y, err = strconv.Atoi(req.y); err != nil {
		errs = append(errs, err)
	}
	if z, err = strconv.Atoi(req.z); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		SetJsonResp(w, http.StatusBadRequest, res.BadRequest("x, y, z are required and should be numeric"))
		return
	}

	if x > 100 || x < 1 || y > 26 || y < 1 || z > 10000 || z < 1 {
		SetJsonResp(w, http.StatusBadRequest, res.BadRequest("x, y, z have wrong values"))
		return
	}

	campaigns := []Campaign{}
	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i1 := 0; i1 < z; i1++  {
		campaign := Campaign{
			Name: fmt.Sprintf("%v%v" , "campaign", i1 + 1),
		}

		targetListLen := r1.Intn(y) // length is random and less than Y //target list всегда будет не пустой

		for i2 := 0; i2 <= targetListLen; i2++ {
			char := string(CHARS[i2])
			attrListLen := r1.Intn(x) 	//	length is random and less than X
			attrList := []string{}
			for i3 := 0; i3 <= attrListLen; i3++ {
				attrList = append(attrList, fmt.Sprintf("%v%v" , char, i3))
			}

			target := CampaignTarget{
				Target: "attr_" + char,
				AttrList: attrList,
			}

			campaign.TargetList = append(campaign.TargetList, target)
		}

		campaigns = append(campaigns, campaign)
	}

	bts, _ := json.Marshal(campaigns)
	if err := ioutil.WriteFile(DATA_PATH+DATA_NAME, bts, DATA_PERM); err != nil {
		SetJsonResp(w, http.StatusInternalServerError, res.BadRequest(err.Error()))
		return
	}
	SetJsonResp(w, http.StatusOK, res.Ok())
}

func main() {
	http.HandleFunc("/campaign", generateCampData)
	port := fmt.Sprintf(":%v", PORT)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func SetJsonResp(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	bts, _ := json.Marshal(data)
	w.Write(bts)
}