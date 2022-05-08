package queries

import (
	"io"

	"github.com/elastic/go-elasticsearch/v7/esutil"
)

const (
	defaultPageSize = 1
)

type dict map[string]interface{}

func SearchCategories(text string) io.Reader {
	var res dict = make(dict)

	res["query"] = dict{
		"bool": dict{
			"should": []dict{
				dict{
					"match": dict{
						"query.name.strict": dict{
							"query":    text,
							"operator": "and",
						},
					},
				},
				dict{
					"match": dict{
						"query.name.fuzzy": dict{
							"query":     text,
							"operator":  "and",
							"fuzziness": 1,
						},
					},
				},
				dict{
					"match": dict{
						"query.name.edge": dict{
							"query": text,
						},
					},
				},
			},
		},
	}

	return esutil.NewJSONReader(res)
}
