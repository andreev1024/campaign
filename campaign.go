package main

type Campaign struct {
	Name       string `json:"compaign_name"`
	Price      float64 `json:"price"`
	TargetList []CampaignTarget `json:"target_list"`
}

type CampaignTarget struct {
	Target   string `json:"target"`
	AttrList []string `json:"attr_list"`
}
