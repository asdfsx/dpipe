#!/bin/sh

# get current template related settings
curl -XGET 'http://localhost:9200/rs/_mapping?pretty=true'

curl -XGET http://localhost:9200/_template?pretty=1

curl -XGET http://localhost:9200/_template/fun?pretty=1

# put template to ES
curl -XPUT localhost:9200/_template/fun -d '
{
    "template": "*",

    "settings": {
        "index": {
            "number_of_shards": 3,
            "number_of_replicas": 0,
            "warmer.enabled": true,
            "refresh_interval": "25s",
            "query" : { "default_field" : "area" }
        }
    },

    "mappings": {
        "_default_": {
            "_source": {
                "enabled": true,
                "compress": true
            }, 
            "_all": {
                "enabled": false
            },
            "_ttl": {
                "enabled": false
            },
	        "_timestamp": {
	            "enabled": true,
	            "path": "t",
                "store": true,
                "index": "not_analyzed"
	        },

            "dynamic_templates": [
                {
                    "string_template" : {
                        "match" : "*",
                        "mapping": { "type": "string", "index": "not_analyzed" },
                        "match_mapping_type" : "string"
                    }
                }
            ],

            "properties" : {
                "area": {
                    "type": "string",
                    "index": "not_analyzed"
                },
                "t": {
                    "type": "date"
                },
                "loc": {
                    "type": "geo_point"
                },
                "typ": {
                    "type": "string",
                    "index": "not_analyzed"
                }
            }
        },

        "dau": {
            "properties": {
                "date": {
                    "type": "date",
                    "format": "YYYYMMdd",
                    "include_in_all": true,
                    "index": "not_analyzed"
                }
            }
        }

    }
}'

