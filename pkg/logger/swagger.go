/*
Package logger provides logging functionality with HTTP endpoints for configuration and log retrieval.

The Swagger/OpenAPI documentation for this package describes two main endpoints:

	/api/loggersettings/debug (POST)
	    Enables or disables debug logging mode. Requires authentication.
	    Example request:
	        POST /api/loggersettings/debug
	        {
	            "enabled": true
	        }

	/api/logging/log (GET/POST)
	    Retrieves log entries with flexible filtering options. Requires authentication.
	    Supports multiple output formats: json, jsonpretty, csv, and text.
	    Example GET request:
	        GET /api/logging/log?last_lines=100&format=json

	    Example POST request:
	        POST /api/logging/log
	        {
	            "from_time": "2024-03-10T15:04:05Z",
	            "last_minutes": 30,
	            "format": "csv"
	        }

Authentication:
The endpoints support three authentication methods:
  - Bearer token (JWT)
  - API Key in header (X-API-Key)
  - API Key in query parameter (API-KEY)

Usage:
To integrate these endpoints into your OpenAPI/Swagger documentation:

	func main() {
	    // Get the logger's Swagger definition
	    loggerSwagger := logger.GetSwagger()

	    // Merge it with your main Swagger documentation
	    mainSwagger := loadMainSwagger()
	    for path, def := range loggerSwagger.Paths {
	        mainSwagger.Paths[path] = def
	    }
	    for name, schema := range loggerSwagger.Components.Schemas {
	        mainSwagger.Components.Schemas[name] = schema
	    }
	}

The endpoints can be tested using curl:

	# Enable debug logging
	curl -X POST http://localhost:8080/api/loggersettings/debug \
	    -H "Authorization: Bearer <your-token>" \
	    -H "Content-Type: application/json" \
	    -d '{"enabled":true}'

	# Get last 100 lines in CSV format
	curl "http://localhost:8080/api/logging/log?last_lines=100&format=csv" \
	    -H "Authorization: Bearer <your-token>"
*/
package logger

import "encoding/json"

// SwaggerDefinition contains the OpenAPI/Swagger paths and schemas for the logger endpoints
type SwaggerDefinition struct {
	Paths      map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components"`
}

// GetSwagger returns the OpenAPI/Swagger definition for the logger endpoints
func GetSwagger() *SwaggerDefinition {
	return &SwaggerDefinition{
		Paths: map[string]interface{}{
			"/api/loggersettings/debug": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Set debug logging mode",
					"tags":    []string{"Logging"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
						{"apiKeyHeader": []string{}},
						{"apiKeyQuery": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/DebugSettings",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Debug settings updated successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/DebugSettings",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Invalid request body",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized - Invalid or missing authentication",
						},
						"405": map[string]interface{}{
							"description": "Method not allowed",
						},
					},
				},
			},
			"/api/logging/log": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Retrieve log entries",
					"tags":    []string{"Logging"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
						{"apiKeyHeader": []string{}},
						{"apiKeyQuery": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "from_time",
							"in":          "query",
							"description": "Start time (RFC3339)",
							"schema": map[string]interface{}{
								"type":   "string",
								"format": "date-time",
							},
						},
						{
							"name":        "to_time",
							"in":          "query",
							"description": "End time (RFC3339)",
							"schema": map[string]interface{}{
								"type":   "string",
								"format": "date-time",
							},
						},
						{
							"name":        "last_lines",
							"in":          "query",
							"description": "Number of recent lines",
							"schema": map[string]interface{}{
								"type":    "integer",
								"minimum": 1,
							},
						},
						{
							"name":        "last_minutes",
							"in":          "query",
							"description": "Number of recent minutes",
							"schema": map[string]interface{}{
								"type":    "integer",
								"minimum": 1,
							},
						},
						{
							"name":        "format",
							"in":          "query",
							"description": "Output format",
							"schema": map[string]interface{}{
								"type":    "string",
								"enum":    []string{"json", "jsonpretty", "csv", "text"},
								"default": "json",
							},
						},
					},
					"responses": getLogResponseDefinition(),
				},
				"post": map[string]interface{}{
					"summary": "Retrieve log entries",
					"tags":    []string{"Logging"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
						{"apiKeyHeader": []string{}},
						{"apiKeyQuery": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"from_time": map[string]interface{}{
											"type":        "string",
											"format":      "date-time",
											"description": "Start time (RFC3339)",
										},
										"to_time": map[string]interface{}{
											"type":        "string",
											"format":      "date-time",
											"description": "End time (RFC3339)",
										},
										"last_lines": map[string]interface{}{
											"type":        "integer",
											"minimum":     1,
											"description": "Number of recent lines",
										},
										"last_minutes": map[string]interface{}{
											"type":        "integer",
											"minimum":     1,
											"description": "Number of recent minutes",
										},
										"format": map[string]interface{}{
											"type":        "string",
											"enum":        []string{"json", "jsonpretty", "csv", "text"},
											"default":     "json",
											"description": "Output format",
										},
									},
								},
							},
						},
					},
					"responses": getLogResponseDefinition(),
				},
			},
		},
		Components: map[string]interface{}{
			"schemas": map[string]interface{}{
				"DebugSettings": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"enabled": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether debug logging is enabled",
						},
					},
					"required": []string{"enabled"},
				},
				"LogResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"lines": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
							"description": "Array of log lines",
						},
					},
				},
			},
		},
	}
}

// Helper function to avoid duplicating the response definition
func getLogResponseDefinition() map[string]interface{} {
	return map[string]interface{}{
		"200": map[string]interface{}{
			"description": "Log entries retrieved successfully",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"$ref": "#/components/schemas/LogResponse",
					},
				},
				"text/csv": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
				"text/plain": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		"400": map[string]interface{}{
			"description": "Invalid parameters",
		},
		"401": map[string]interface{}{
			"description": "Unauthorized - Invalid or missing authentication",
		},
		"405": map[string]interface{}{
			"description": "Method not allowed",
		},
		"500": map[string]interface{}{
			"description": "Internal server error",
		},
	}
}

// GetSwaggerJSON returns the OpenAPI/Swagger definition as a JSON string
func GetSwaggerJSON() (string, error) {
	swagger := GetSwagger()
	data, err := json.MarshalIndent(swagger, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
