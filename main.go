/*
	generate data
		http GET http://localhost:3000/campaign x==100 y==26 z==10000 -v

	import campaign data
		http POST http://localhost:3000/import_camp

	search
		http POST http://localhost:3000/search user=u1 profile:='{"attr_A":"A5","attr_B":"B15", "attr_C":"C15", "attr_D":"D10", "attr_E":"E10"}' -v -j

*/
package main

import (
	"campaign/res"
	"campaign/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PORT  = 3000
	CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	DATA_PERM = 0644
	DATA_NAME = "data.json"
	DATA_PATH = "./"

	MAX_WORKER_LIMIT = 10000
)

var campaigns []Campaign
var mutex sync.Mutex

//@todo singletone
func Rand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func generateCampDataHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		SetJsonResp(w, http.StatusMethodNotAllowed, res.MethodNotAllowed())
		return
	}

	start := time.Now()

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

	campaigns := generateCampData(x, y, z)

	bts, _ := json.Marshal(campaigns)
	if err := ioutil.WriteFile(DATA_PATH+DATA_NAME, bts, DATA_PERM); err != nil {
		SetJsonResp(w, http.StatusInternalServerError, res.BadRequest(err.Error()))
		return
	}
	SetJsonResp(w, http.StatusOK, res.Ok())

	secs := time.Since(start).Seconds()
	log.Printf("%.2fs", secs)
}

func generateCampData(x, y, z int) (campaigns []Campaign) {
	rand := Rand()
	for i1 := 0; i1 < z; i1++ {
		campaign := Campaign{
			Name:  fmt.Sprintf("%v%v", "campaign", i1+1),
			Price: utils.TruncateFloat(Rand().Float64()*100, 2),
		}

		targetListLen := rand.Intn(y) // length is random and less than Y //target list всегда будет не пустой

		for i2 := 0; i2 <= targetListLen; i2++ {
			char := string(CHARS[i2])
			attrListLen := rand.Intn(x) //	length is random and less than X
			attrList := []string{}
			for i3 := 0; i3 <= attrListLen; i3++ {
				attrList = append(attrList, fmt.Sprintf("%v%v", char, i3))
			}

			target := CampaignTarget{
				Target:   "attr_" + char,
				AttrList: attrList,
			}

			campaign.TargetList = append(campaign.TargetList, target)
		}
		campaigns = append(campaigns, campaign)
	}
	return
}

func importCampHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	status, msg := importCamp()
	SetJsonResp(w, status, msg)

	secs := time.Since(start).Seconds()
	log.Printf("%.2fs", secs)
}

func importCamp() (status int, msg interface{}) {
	status = http.StatusOK
	msg = res.Ok()
	var data []byte
	var err error
	if data, err = ioutil.ReadFile(DATA_PATH + DATA_NAME); err != nil {
		status = http.StatusInternalServerError
		msg = res.InternalServerError(err.Error())
		return
	}
	if err = json.Unmarshal(data, &campaigns); err != nil {
		status = http.StatusInternalServerError
		msg = res.InternalServerError(err.Error())
		return
	}
	return
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	if r.Method != http.MethodPost {
		SetJsonResp(w, http.StatusMethodNotAllowed, res.MethodNotAllowed())
		return
	}

	if len(campaigns) == 0 {
		SetJsonResp(w, http.StatusBadRequest, res.BadRequest("Please, generate and import data"))
		return
	}

	var contentType string
	var u User
	if ct := r.Header.Get("Content-Type"); ct != "" {
		contentType = strings.Split(ct, ";")[0]
	}

	switch contentType {
	case "application/json":
		var bts []byte
		var err error
		if bts, err = ioutil.ReadAll(r.Body); err != nil {
			SetJsonResp(w, http.StatusBadRequest, res.BadRequest(err.Error()))
			return
		}
		if err = json.Unmarshal(bts, &u); err != nil {
			SetJsonResp(w, http.StatusBadRequest, res.BadRequest(err.Error()))
			return
		}
	default:
		SetJsonResp(w, http.StatusBadRequest, res.BadRequest("Content type should be `application/json`"))
		return
	}

	name := search(&u, &campaigns)

	secs := time.Since(start).Seconds()
	log.Printf("%.2fs", secs)

	if utils.IsEmptyStr(name) {
		SetJsonResp(w, http.StatusOK, res.Ok())
		return
	}

	SetJsonResp(w, http.StatusOK, struct {
		Winner string
		Counter int
	}{Winner:name})
	return
}

func search(u *User, c *[]Campaign) (name string) {
	var wg sync.WaitGroup

	taskCount := len(*c)
	workers := MAX_WORKER_LIMIT

	if taskCount < workers {
		workers = taskCount
	}

	log.Println("workers count ", workers)

	wg.Add(workers)

	tasksCh := make(chan Campaign)

	var price float64
	for i := 0; i < workers; i++ {
		go searchRoutine(tasksCh, &wg, u, &price, &name)
	}

	for i := 0; i < taskCount; i++ {
		tasksCh <- (*c)[i]
	}

	close(tasksCh)
	wg.Wait()
	return
}

func searchRoutine(tasksCh <-chan Campaign, wg *sync.WaitGroup, u *User, price *float64, name *string) {
	defer wg.Done()
A:
	for {
		campaign, ok := <-tasksCh
		if !ok {
			return
		}

		//If all targets of the campaign can be found in an user's profile, and
		for _, ct := range campaign.TargetList {
			if _, ok := u.Profile[ct.Target]; ok {
				continue
			}
		}

		//the user’s profile attribute value can be found in the list of the campaign target attri_list.
	B:
		for _, ct := range campaign.TargetList { // B
			attrValue := u.Profile[ct.Target]
			for _, a := range ct.AttrList {
				if attrValue == a {
					continue B
				}
			}
			continue A
		}

		mutex.Lock()
		if *price < campaign.Price || utils.IsEmptyStr(*name) {
			*price = campaign.Price
			*name = campaign.Name
		}
		mutex.Unlock()
	}
}

func main() {
	http.HandleFunc("/campaign", generateCampDataHandler)
	http.HandleFunc("/import_camp", importCampHandler)
	http.HandleFunc("/search", searchHandler)

	port := fmt.Sprintf(":%v", PORT)

	log.SetFlags(log.LstdFlags | log.Lshortfile)

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