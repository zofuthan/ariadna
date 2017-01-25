package search

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gen1us2k/go-translit"
	"github.com/gen1us2k/log"
	"github.com/maddevsio/ariadna/common"
	"github.com/maddevsio/ariadna/config"
	"github.com/maddevsio/ariadna/geo"
	"github.com/maddevsio/ariadna/models"
	"gopkg.in/olivere/elastic.v3"
)

type Search interface {
	GetCurrentIndexName() (string, error)
}

type ElasticSearch struct {
	Search
	client    *elastic.Client
	logger    log.Logger
	appConfig *config.AriadnaConfig
}

func NewElasticSearch(conf *config.AriadnaConfig) (*ElasticSearch, error) {
	e := &ElasticSearch{}
	client, err := elastic.NewClient(
		elastic.SetURL(conf.ElasticSearchHost),
	)
	if err != nil {
		return nil, err
	}
	e.client = client
	e.logger = log.NewLogger("elasticsearch")
	e.appConfig = conf
	return e, nil
}

func (es *ElasticSearch) GetCurrentIndexName() (string, error) {
	// TODO: remove hardcoded value
	indexName := ""
	res, err := es.client.Aliases().Index("_all").Do()
	if err != nil {
		return "", err
	}
	for _, index := range res.IndicesByAlias(indexName) {
		if strings.HasPrefix(index, indexName) {
			return index, nil
		}
	}
	return "", nil
}

//func (es *ElasticSearch) JsonWaysToES(Addresses []models.JsonWay, CitiesAndTowns []models.JsonWay, client *elastic.Client) {
//	es.logger.Info("Populating elastic search index")
//	bulkClient := client.Bulk()
//	es.logger.Info("Creating bulk client")
//	for _, address := range Addresses {
//		cityName, villageName, suburbName, townName := "", "", "", ""
//		var lat, _ = strconv.ParseFloat(address.Centroid["lat"], 64)
//		var lng, _ = strconv.ParseFloat(address.Centroid["lon"], 64)
//		for _, city := range CitiesAndTowns {
//			polygon := geo.NewPolygon(city.Nodes)
//
//			if polygon.Contains(geo.NewPoint(lat, lng)) {
//				switch city.Tags["place"] {
//				case "city":
//					cityName = city.Tags["name"]
//				case "village":
//					villageName = city.Tags["name"]
//				case "suburb":
//					suburbName = city.Tags["name"]
//				case "town":
//					townName = city.Tags["name"]
//				case "neighbourhood":
//					suburbName = city.Tags["name"]
//				}
//			}
//		}
//		var points [][][]float64
//		for _, point := range address.Nodes {
//			points = append(points, [][]float64{[]float64{point.Lat(), point.Lon()}})
//		}
//
//		pg := gj.NewPolygonFeature(points)
//		centroid := make(map[string]float64)
//		centroid["lat"] = lat
//		centroid["lon"] = lng
//		name := common.CleanAddress(address.Tags["name"])
//		translated := ""
//
//		if latinre.Match([]byte(name)) {
//			word := make(map[string]string)
//			word["original"] = name
//
//			trans := strings.Split(name, " ")
//			for _, k := range trans {
//				s := synonims[k]
//				if s == "" {
//					s = translit.Translit(k)
//				}
//				translated += fmt.Sprintf("%s ", s)
//			}
//
//			word["trans"] = translated
//		}
//		housenumber := translit.Translit(address.Tags["addr:housenumber"])
//		marshall := models.JsonEsIndex{
//			Country:           "KG",
//			City:              cityName,
//			Village:           villageName,
//			Town:              townName,
//			District:          suburbName,
//			Street:            common.CleanAddress(address.Tags["addr:street"]),
//			HouseNumber:       housenumber,
//			Name:              name,
//			OldName:           address.Tags["old_name"],
//			HouseName:         address.Tags["housename"],
//			PostCode:          address.Tags["postcode"],
//			LocalName:         address.Tags["loc_name"],
//			AlternativeName:   address.Tags["alt_name"],
//			InternationalName: address.Tags["int_name"],
//			NationalName:      address.Tags["nat_name"],
//			OfficialName:      address.Tags["official_name"],
//			RegionalName:      address.Tags["reg_name"],
//			ShortName:         address.Tags["short_name"],
//			SortingName:       address.Tags["sorting_name"],
//			TranslatedName:    translated,
//			Centroid:          centroid,
//			Geom:              pg,
//			Custom:            false,
//		}
//		index := elastic.NewBulkIndexRequest().
//			Index(common.AC.ElasticSearchIndexUrl).
//			Type(common.AC.IndexType).
//			Id(strconv.FormatInt(address.ID, 10)).
//			Doc(marshall)
//		bulkClient = bulkClient.Add(index)
//	}
//	es.logger.Info("Starting to insert many data to elasticsearch")
//	_, err := bulkClient.Do()
//	if err != nil {
//		es.logger.Error(err.Error())
//	}
//}

//func (es *ElasticSearch) AddJSONNode()
func (es *ElasticSearch) JsonNodesToEs(Addresses []models.JsonNode, CitiesAndTowns []models.JsonWay, client *elastic.Client) {
	bulkClient := client.Bulk()
	for _, address := range Addresses {
		cityName, villageName, suburbName, townName := "", "", "", ""
		for _, city := range CitiesAndTowns {
			polygon := geo.NewPolygon(city.Nodes)

			if polygon.Contains(geo.NewPoint(address.Lat, address.Lon)) {
				switch city.Tags["place"] {
				case "city":
					cityName = city.Tags["name"]
				case "village":
					villageName = city.Tags["name"]
				case "suburb":
					suburbName = city.Tags["name"]
				case "town":
					townName = city.Tags["name"]
				case "neighbourhood":
					suburbName = city.Tags["name"]
				}
			}
		}

		centroid := make(map[string]float64)
		centroid["lat"] = address.Lat
		centroid["lon"] = address.Lon
		name := common.CleanAddress(address.Tags["name"])
		translated := ""
		if common.Latinre.Match([]byte(name)) {
			word := make(map[string]string)
			word["original"] = name

			trans := strings.Split(name, " ")
			for _, k := range trans {
				s := common.Synonims[k]
				if s == "" {
					s = translit.Translit(k)
				}
				translated += fmt.Sprintf("%s ", s)
			}

			word["trans"] = translated
		}
		housenumber := translit.Translit(address.Tags["addr:housenumber"])

		marshall := models.JsonEsIndex{
			Country:           "KG",
			City:              cityName,
			Village:           villageName,
			Town:              townName,
			District:          suburbName,
			Street:            common.CleanAddress(address.Tags["addr:street"]),
			HouseNumber:       housenumber,
			Name:              name,
			TranslatedName:    translated,
			OldName:           address.Tags["old_name"],
			HouseName:         address.Tags["housename"],
			PostCode:          address.Tags["postcode"],
			LocalName:         address.Tags["loc_name"],
			AlternativeName:   address.Tags["alt_name"],
			InternationalName: address.Tags["int_name"],
			NationalName:      address.Tags["nat_name"],
			OfficialName:      address.Tags["official_name"],
			RegionalName:      address.Tags["reg_name"],
			ShortName:         address.Tags["short_name"],
			SortingName:       address.Tags["sorting_name"],
			Centroid:          centroid,
			Geom:              nil,
			Custom:            false,
			Intersection:      address.Intersection,
		}

		index := elastic.NewBulkIndexRequest().
			Index(es.appConfig.ElasticSearchIndexUrl).
			Type(es.appConfig.IndexType).
			Id(strconv.FormatInt(address.ID, 10)).
			Doc(marshall)
		bulkClient = bulkClient.Add(index)
	}
	_, err := bulkClient.Do()
	if err != nil {
		es.logger.Error(err.Error())
	}

}