package locations

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func GetGaodeLocation(lat, long float64, apiKey string) *NominatimLocation {
	// Add throttling
	if time.Since(lastRequest) < throttling {
		time.Sleep(throttling - time.Since(lastRequest))
	}
	lastRequest = time.Now()

	url := fmt.Sprintf("https://restapi.amap.com/v3/geocode/regeo?key=%s&location=%f,%f&extensions=all&batch=false&roadlevel=0&output=JSON", apiKey, long, lat)
	log.Printf("Making request to: %s", url)
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed request to:", url, err)
		return nil
	}
	defer resp.Body.Close()

	var gaodeResp struct {
		Regeocode struct {
			FormattedAddress string `json:"formatted_address"`
			AddressComponent struct {
				City        []string `json:"city"`
				Province    string   `json:"province"`
				District    string   `json:"district"`
				Township    string   `json:"township"`
				Country     string   `json:"country"`
				CountryCode string   `json:"country_code"`
			} `json:"addressComponent"`
		} `json:"regeocode"`
	}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&gaodeResp); err != nil {
		log.Println(url, err)
		return nil
	}
	result := &NominatimLocation{
		DisplayName: gaodeResp.Regeocode.FormattedAddress,
		Address: NominatimAddress{
			City:          "",
			Province:      gaodeResp.Regeocode.AddressComponent.Province,
			Neighbourhood: gaodeResp.Regeocode.AddressComponent.Township,
			Country:       gaodeResp.Regeocode.AddressComponent.Country,
			CountryCode:   gaodeResp.Regeocode.AddressComponent.CountryCode,
		},
	}
	if len(gaodeResp.Regeocode.AddressComponent.City) > 0 {
		result.Address.City = gaodeResp.Regeocode.AddressComponent.City[0]
	} else {
		result.Address.City = gaodeResp.Regeocode.AddressComponent.Province
	}
	if result.DisplayName == "" {
		// Build a display name manually
		parts := []string{}
		if result.Address.Neighbourhood != "" {
			parts = append(parts, result.Address.Neighbourhood)
		}
		if result.Address.City != "" {
			parts = append(parts, result.Address.City)
		}
		if result.Address.Province != "" {
			parts = append(parts, result.Address.Province)
		}
		if result.Address.Country != "" {
			parts = append(parts, result.Address.Country)
		}
		result.DisplayName = strings.Join(parts, ", ")
	}
	return result
}
