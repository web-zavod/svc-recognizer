package queries

import (
	"io"
	"strings"
)

// Returns the Elasticsearch request to create the category index
func CreateCategoriesIndex() io.Reader {
	return strings.NewReader(`{
  "settings": {
    "index": {
      "max_ngram_diff": 30
    },
    "analysis": {
      "filter": {
        "ngram": {
          "type": "ngram",
          "min_gram": 1,
          "max_gram": 30
        },
        "edge_ngram": {
          "type": "edge_ngram",
          "min_gram": 1,
          "max_gram": 20
        },
        "name_remove_noise": {
          "type": "pattern_replace",
          "pattern": ".*[^а-яa-z].*",
          "replace": ""
        },
        "name_length": {
          "type": "length",
          "min": 3
        },
        "russian_stop": {
          "type": "stop",
          "stopwords": "_russian_"
        },
        "russian_stemmer": {
          "type": "stemmer",
          "language": "russian"
        }
      },
      "analyzer": {
        "index_name_edge": {
          "type": "custom",
          "tokenizer": "keyword",
          "filter": [
            "lowercase",
            "name_remove_noise",
            "name_length",
            "edge_ngram"
          ]
        },
        "index_name_strict": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "lowercase",
            "name_remove_noise",
            "name_length",
            "russian_stop"
          ]
        },
        "index_name_fuzzy": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "lowercase",
            "name_remove_noise",
            "name_length",
            "russian_stop",
            "russian_stemmer",
            "edge_ngram"
          ]
        },
        "search_keywords": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "lowercase"
          ]
        },
        "search_full_text": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "lowercase",
            "russian_stop",
            "russian_stemmer"
          ]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": {
        "type": "keyword"
      },
      "query": {
        "properties": {
          "name": {
            "type": "keyword",
            "fields": {
              "edge": {
                "type": "text",
                "analyzer": "index_name_edge",
                "search_analyzer": "search_keywords"
              },
              "strict": {
                "type": "text",
                "analyzer": "index_name_strict",
                "search_analyzer": "search_keywords"
              },
              "fuzzy": {
                "type": "text",
                "analyzer": "index_name_fuzzy",
                "search_analyzer": "search_full_text"
              }
            }
          }
        }
      }
    }
  }
}`)
}
