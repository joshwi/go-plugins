package graphdb

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"

	"github.com/joshwi/go-utils/logger"
	"github.com/joshwi/go-utils/parser"
	"github.com/joshwi/go-utils/utils"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

var regexp_1 = regexp.MustCompile(`"`)

func Connect(url string, username string, password string) neo4j.Driver {

	Neo4jConfig := func(conf *neo4j.Config) { conf.Encrypted = false }

	driver, err := neo4j.NewDriver(url, neo4j.BasicAuth(username, password, ""), Neo4jConfig)
	if err != nil {
		log.Println(err)
	}

	return driver
}

func RunCypher(session neo4j.Session, query string) [][]utils.Tag {

	output := [][]utils.Tag{}

	// defer session.Close()

	result, err := session.Run(query, map[string]interface{}{})
	if err != nil {
		log.Println(err)
	}

	for result.Next() {
		entry := []utils.Tag{}
		keys := result.Record().Keys()
		for n := 0; n < len(keys); n++ {
			value := fmt.Sprintf("%v", result.Record().GetByIndex(n))
			input := utils.Tag{
				Name:  keys[n],
				Value: value,
			}
			entry = append(entry, input)
		}
		output = append(output, entry)
	}

	return output
}

func PostNode(session neo4j.Session, node string, label string, properties []utils.Tag) string {

	cypher := `CREATE (n: ` + node + ` { label: "` + label + `" })`

	for _, item := range properties {
		cypher += ` SET n.` + item.Name + ` = "` + regexp_1.ReplaceAllString(item.Value, "\\'") + `"`
	}

	// cypher = regexp_1.ReplaceAllString(cypher, "'")

	result, err := session.Run(cypher, map[string]interface{}{})
	if err != nil {
		log.Println(err)
	}

	summary, err := result.Summary()

	counters := summary.Counters()

	output := fmt.Sprintf(`[ Function: PutNode ] [ Label: %v ] [ Node: %v ] [ Properties Set: %v ]`, label, node, counters.PropertiesSet())

	return output
}

func PutNode(session neo4j.Session, node string, label string, properties []utils.Tag) string {

	cypher := `MERGE (n: ` + node + ` { label: "` + label + `" })`

	for _, item := range properties {
		cypher += ` SET n.` + item.Name + ` = "` + regexp_1.ReplaceAllString(item.Value, "\\'") + `"`
	}

	result, err := session.Run(cypher, map[string]interface{}{})
	if err != nil {
		log.Println(err)
	}

	summary, err := result.Summary()
	if err != nil {
		log.Println(err)
	}

	counters := summary.Counters()

	output := fmt.Sprintf(`[ Function: PutNode ] [ Label: %v ] [ Node: %v ] [ Properties Set: %v ]`, label, node, counters.PropertiesSet())

	return output

}

// func DeleteNode(driver string, node string, label string) {
// }

func StoreDB(driver neo4j.Driver, params map[string]string, label string, bucket string, data utils.Output, wg *sync.WaitGroup) {

	count := []string{}

	defer wg.Done()

	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		log.Println(err)
	}

	for _, item := range data.Collections {
		for n, entry := range item.Value {
			properties := []utils.Tag{}
			properties = append(properties, data.Tags...)
			properties = append(properties, entry...)
			new_bucket := bucket + "_" + item.Name
			new_label := label + "_" + strconv.Itoa(n+1)
			text := PutNode(session, new_bucket, new_label, properties)
			count = append(count, text)
		}
	}

	logger.Logger.Info().Str("collector", bucket).Str("query", fmt.Sprintf("%v", params)).Int("nodes", len(count)).Msg("Store DB")

	session.Close()
}

func RunScript(driver neo4j.Driver, entry []utils.Tag, config utils.Config, wg *sync.WaitGroup) {

	// Convert params from struct [][]utils.Tag -> map[string]string
	params := map[string]string{}

	for _, item := range entry {
		params[item.Name] = item.Value
	}

	// Add params to search urls
	urls := parser.AddParams(params, config.Urls, config.Params)

	// Run GET request and parsing collection
	label, bucket, data := parser.RunJob(params, urls, config)

	// Send output data to Neo4j
	StoreDB(driver, params, label, bucket, data, wg)

}
