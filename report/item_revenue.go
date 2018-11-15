package report

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/TerrexTech/go-mongoutils/mongo"
	"github.com/mongodb/mongo-go-driver/bson"
	mgo "github.com/mongodb/mongo-go-driver/mongo"
	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomFloat(min, max float64) float64 {
	return rand.Float64() * (max - min)
}

func generateRandomFloat(num1, num2 float64) float64 {
	// rand.Seed(time.Now().Unix())
	return randomFloat(num1, num2)
}

func ItemSoldReport(aggParams SoldItemParams, itemSoldColl *mongo.Collection) ([]interface{}, error) {

	if aggParams.Timestamp.Lt == 0 || aggParams.Timestamp.Gt == 0 {
		err := errors.New("Missing timestamp value")
		log.Println(err)
		return nil, err
	}
	input, err := json.Marshal(aggParams)
	if err != nil {
		err = errors.Wrap(err, "Unable to marshal aggParams")
		log.Println(err)
		return nil, err
	}

	log.Println(input)
	log.Println(aggParams)

	pipelineBuilder := fmt.Sprintf(`[
		{
			"$match": %s
		},
		{
			"$group" : {
			"_id" : {"sku" : "$sku","name":"$name"},
			"avg_sold": {
				"$avg": "$weight",
			}
		}
		}
	]`, input)

	// ,
	// 		"avg_total": {
	// 			"$avg": "$totalWeight",
	// 		}

	pipelineAgg, err := bson.ParseExtJSONArray(pipelineBuilder)
	if err != nil {
		err = errors.Wrap(err, "Query: Error in generating pipeline for report")
		log.Println(err)
		return nil, err
	}

	findResult, err := itemSoldColl.Aggregate(pipelineAgg)
	if err != nil {
		err = errors.Wrap(err, "Query: Error in getting aggregate results ")
		log.Println(err)
		return nil, err
	}
	return findResult, nil
}

func RevenueSoldWeight(avgSoldReport []interface{}) []ReportResult {
	var reportAgg []ReportResult

	for _, v := range avgSoldReport {
		m, assertOK := v.(map[string]interface{})
		if !assertOK {
			err := errors.New("Error getting results ")
			log.Fatalln(err)
		}

		groupByFields := m["_id"]
		mapInGroupBy := groupByFields.(map[string]interface{})
		sku := mapInGroupBy["sku"].(string)
		name := mapInGroupBy["name"].(string)

		//Generate value for previous year
		currSoldWeight := m["avg_sold"].(float64)
		prevSoldWeight := currSoldWeight / generateRandomFloat(0.1, 2.8)

		revenueCurrRandPrice := generateRandomFloat(0.5, 5.9)
		revenueCurr := currSoldWeight * revenueCurrRandPrice

		revenuePrev := prevSoldWeight * generateRandomFloat(0.1, revenueCurrRandPrice)
		revenuePercent := ((revenueCurr - revenuePrev) / revenuePrev) * 100

		reportAgg = append(reportAgg, ReportResult{
			SKU:            sku,
			Name:           name,
			SoldWeight:     currSoldWeight,
			PrevSoldWeight: prevSoldWeight,
			RevenuePrev:    revenuePrev,
			RevenueCurr:    revenueCurr,
			RevenuePercent: revenuePercent,
		})
		// reportAgg = []ReportResult{
		// 	ReportResult{
		// 		SKU:            sku,
		// 		Name:           name,
		// 		SoldWeight:     currSoldWeight,
		// 		PrevSoldWeight: prevSoldWeight,
		// 		RevenuePrev:    revenuePrev,
		// 		RevenueCurr:    revenueCurr,
		// 		RevenuePercent: revenuePercent,
		// 	},
		// }
	}
	return reportAgg
}

func CreateReport(reportGen SoldReport, reportColl *mongo.Collection) (*mgo.InsertOneResult, error) {
	insertRep, err := reportColl.InsertOne(reportGen)
	if err != nil {
		err = errors.Wrap(err, "Query: Error in generating report ")
		log.Println(err)
		return nil, err
	}
	return insertRep, nil
}
