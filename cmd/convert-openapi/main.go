package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: convert-openapi <input-file> <output-file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Read input file
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal("Failed to read input file:", err)
	}

	// Parse as Swagger 2.0
	var swagger map[string]interface{}
	if err := json.Unmarshal(data, &swagger); err != nil {
		// Try YAML if JSON fails
		if err := yaml.Unmarshal(data, &swagger); err != nil {
			log.Fatal("Failed to parse input file:", err)
		}
	}

	// Convert to OpenAPI 3.0
	openapi := convertSwaggerToOpenAPI(swagger)

	// Write output file
	var output []byte
	if outputFile[len(outputFile)-4:] == "yaml" || outputFile[len(outputFile)-3:] == "yml" {
		output, err = yaml.Marshal(openapi)
	} else {
		output, err = json.MarshalIndent(openapi, "", "  ")
	}
	if err != nil {
		log.Fatal("Failed to marshal output:", err)
	}

	if err := ioutil.WriteFile(outputFile, output, 0644); err != nil {
		log.Fatal("Failed to write output file:", err)
	}

	fmt.Printf("Successfully converted %s to OpenAPI 3.0 format: %s\n", inputFile, outputFile)
}

func convertSwaggerToOpenAPI(swagger map[string]interface{}) map[string]interface{} {
	openapi := make(map[string]interface{})

	// OpenAPI version
	openapi["openapi"] = "3.0.3"

	// Info
	if info, ok := swagger["info"]; ok {
		openapi["info"] = info
	}

	// Servers (convert from host/basePath/schemes)
	servers := []map[string]interface{}{}
	host := "localhost:8080"
	if h, ok := swagger["host"].(string); ok {
		host = h
	}
	basePath := "/"
	if bp, ok := swagger["basePath"].(string); ok {
		basePath = bp
	}
	schemes := []string{"http"}
	if s, ok := swagger["schemes"].([]interface{}); ok {
		schemes = []string{}
		for _, scheme := range s {
			if str, ok := scheme.(string); ok {
				schemes = append(schemes, str)
			}
		}
	}
	if len(schemes) == 0 {
		schemes = []string{"http"}
	}

	for _, scheme := range schemes {
		servers = append(servers, map[string]interface{}{
			"url": fmt.Sprintf("%s://%s%s", scheme, host, basePath),
		})
	}
	openapi["servers"] = servers

	// Components (convert definitions and securityDefinitions)
	components := make(map[string]interface{})

	// Convert definitions to schemas
	if definitions, ok := swagger["definitions"].(map[string]interface{}); ok {
		schemas := make(map[string]interface{})
		for name, def := range definitions {
			schemas[name] = def
		}
		components["schemas"] = schemas
	}

	// Convert securityDefinitions to securitySchemes
	if secDefs, ok := swagger["securityDefinitions"].(map[string]interface{}); ok {
		secSchemes := make(map[string]interface{})
		for name, secDef := range secDefs {
			if def, ok := secDef.(map[string]interface{}); ok {
				secScheme := make(map[string]interface{})
				if t, ok := def["type"].(string); ok {
					if t == "apiKey" {
						secScheme["type"] = "apiKey"
						if in, ok := def["in"].(string); ok {
							secScheme["in"] = in
						}
						if n, ok := def["name"].(string); ok {
							secScheme["name"] = n
						}
					} else if t == "basic" {
						secScheme["type"] = "http"
						secScheme["scheme"] = "basic"
					} else if t == "oauth2" {
						secScheme["type"] = "oauth2"
						// Add OAuth2 flows conversion if needed
					}
				}
				secSchemes[name] = secScheme
			}
		}
		components["securitySchemes"] = secSchemes
	}

	openapi["components"] = components

	// Paths
	if paths, ok := swagger["paths"].(map[string]interface{}); ok {
		newPaths := make(map[string]interface{})
		for path, pathItem := range paths {
			if pi, ok := pathItem.(map[string]interface{}); ok {
				newPathItem := make(map[string]interface{})
				for method, operation := range pi {
					if op, ok := operation.(map[string]interface{}); ok {
						newOp := convertOperation(op)
						newPathItem[method] = newOp
					}
				}
				newPaths[path] = newPathItem
			}
		}
		openapi["paths"] = newPaths
	}

	// Tags
	if tags, ok := swagger["tags"]; ok {
		openapi["tags"] = tags
	}

	return openapi
}

func convertOperation(op map[string]interface{}) map[string]interface{} {
	newOp := make(map[string]interface{})

	// Copy simple fields
	simpleFields := []string{"summary", "description", "operationId", "tags", "security", "deprecated"}
	for _, field := range simpleFields {
		if val, ok := op[field]; ok {
			newOp[field] = val
		}
	}

	// Convert parameters and requestBody
	if params, ok := op["parameters"].([]interface{}); ok {
		pathParams := []interface{}{}
		queryParams := []interface{}{}
		headerParams := []interface{}{}
		var requestBody map[string]interface{}

		for _, param := range params {
			if p, ok := param.(map[string]interface{}); ok {
				if in, ok := p["in"].(string); ok {
					if in == "body" {
						// Convert body parameter to requestBody
						requestBody = map[string]interface{}{
							"required": p["required"],
						}
						if desc, ok := p["description"]; ok {
							requestBody["description"] = desc
						}
						content := make(map[string]interface{})

						// Check consumes for media types
						mediaType := "application/json"
						if consumes, ok := op["consumes"].([]interface{}); ok && len(consumes) > 0 {
							if mt, ok := consumes[0].(string); ok {
								mediaType = mt
							}
						}

						content[mediaType] = map[string]interface{}{
							"schema": p["schema"],
						}
						requestBody["content"] = content
					} else if in == "path" {
						pathParams = append(pathParams, param)
					} else if in == "query" {
						queryParams = append(queryParams, param)
					} else if in == "header" {
						headerParams = append(headerParams, param)
					}
				}
			}
		}

		// Add non-body parameters
		allParams := append(pathParams, queryParams...)
		allParams = append(allParams, headerParams...)
		if len(allParams) > 0 {
			newOp["parameters"] = allParams
		}
		if requestBody != nil {
			newOp["requestBody"] = requestBody
		}
	}

	// Convert responses
	if responses, ok := op["responses"].(map[string]interface{}); ok {
		newResponses := make(map[string]interface{})
		for code, response := range responses {
			if resp, ok := response.(map[string]interface{}); ok {
				newResp := make(map[string]interface{})
				if desc, ok := resp["description"]; ok {
					newResp["description"] = desc
				} else {
					newResp["description"] = "Response"
				}

				if schema, ok := resp["schema"]; ok {
					content := make(map[string]interface{})

					// Check produces for media types
					mediaType := "application/json"
					if produces, ok := op["produces"].([]interface{}); ok && len(produces) > 0 {
						if mt, ok := produces[0].(string); ok {
							mediaType = mt
						}
					}

					content[mediaType] = map[string]interface{}{
						"schema": schema,
					}
					newResp["content"] = content
				}
				newResponses[code] = newResp
			}
		}
		newOp["responses"] = newResponses
	}

	return newOp
}