// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/schedules": {
            "get": {
                "description": "Get a paginated list of schedules",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schedules"
                ],
                "summary": "List all schedules",
                "parameters": [
                    {
                        "type": "integer",
                        "default": 1,
                        "description": "Page number",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "default": 10,
                        "description": "Items per page",
                        "name": "page_size",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ListSchedulesResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new event schedule with blocks and items",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schedules"
                ],
                "summary": "Create a new schedule",
                "parameters": [
                    {
                        "description": "Schedule object",
                        "name": "schedule",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/schedules/analyze": {
            "post": {
                "description": "Analyze potential risks and get recommendations for a schedule",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schedules"
                ],
                "summary": "Analyze schedule risks",
                "parameters": [
                    {
                        "description": "Schedule to analyze",
                        "name": "schedule",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.AnalysisResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/schedules/optimize": {
            "post": {
                "description": "Optimize a schedule to minimize risks and improve efficiency",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schedules"
                ],
                "summary": "Optimize schedule",
                "parameters": [
                    {
                        "description": "Schedule to optimize",
                        "name": "schedule",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.OptimizationResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/schedules/{id}": {
            "get": {
                "description": "Get detailed information about a specific schedule",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schedules"
                ],
                "summary": "Get a schedule by ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Schedule ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    }
                }
            },
            "put": {
                "description": "Update an existing schedule's information",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schedules"
                ],
                "summary": "Update a schedule",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Schedule ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Updated schedule object",
                        "name": "schedule",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete a schedule by its ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schedules"
                ],
                "summary": "Delete a schedule",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Schedule ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/schedules/{id}/public": {
            "get": {
                "description": "Get a formatted public version of a schedule",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schedules"
                ],
                "summary": "Get public schedule",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Schedule ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "enum": [
                            "json",
                            "text"
                        ],
                        "type": "string",
                        "default": "json",
                        "description": "Output format (json or text)",
                        "name": "format",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "When format=text",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/schedules/{id}/volunteer": {
            "get": {
                "description": "Get a formatted version of a schedule for volunteers",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schedules"
                ],
                "summary": "Get volunteer schedule",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Schedule ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.VolunteerSchedule"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "cor-events-scheduler_internal_domain_models.Block": {
            "type": "object",
            "required": [
                "duration",
                "name"
            ],
            "properties": {
                "complexity": {
                    "type": "number",
                    "example": 0.7
                },
                "created_at": {
                    "type": "string"
                },
                "dependencies": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "duration": {
                    "type": "integer",
                    "minimum": 1,
                    "example": 120
                },
                "equipment": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Equipment"
                    }
                },
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.BlockItem"
                    }
                },
                "location": {
                    "type": "string",
                    "example": "Main Stage"
                },
                "max_participants": {
                    "type": "integer",
                    "example": 1000
                },
                "name": {
                    "type": "string",
                    "example": "Main Stage Performance"
                },
                "order": {
                    "type": "integer",
                    "example": 1
                },
                "required_staff": {
                    "type": "integer",
                    "example": 10
                },
                "risk_factors": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.RiskFactor"
                    }
                },
                "schedule_id": {
                    "type": "integer",
                    "example": 1
                },
                "start_time": {
                    "type": "string",
                    "example": "2024-07-01T14:00:00Z"
                },
                "tech_break_duration": {
                    "type": "integer",
                    "example": 30
                },
                "type": {
                    "type": "string",
                    "example": "performance"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.BlockItem": {
            "type": "object",
            "properties": {
                "block_id": {
                    "type": "integer",
                    "example": 1
                },
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string",
                    "example": "Live performance by the main band"
                },
                "duration": {
                    "type": "integer",
                    "example": 45
                },
                "equipment": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Equipment"
                    }
                },
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "name": {
                    "type": "string",
                    "example": "Band Performance"
                },
                "order": {
                    "type": "integer",
                    "example": 1
                },
                "participants": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Participant"
                    }
                },
                "requirements": {
                    "type": "string",
                    "example": "Stage lighting, sound system"
                },
                "type": {
                    "type": "string",
                    "example": "performance"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.Equipment": {
            "type": "object",
            "properties": {
                "complexity_score": {
                    "type": "number",
                    "example": 0.8
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "name": {
                    "type": "string",
                    "example": "Sound System"
                },
                "setup_time": {
                    "type": "integer",
                    "example": 30
                },
                "type": {
                    "type": "string",
                    "example": "audio"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.Participant": {
            "type": "object",
            "properties": {
                "block_item_id": {
                    "type": "integer",
                    "example": 1
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "name": {
                    "type": "string",
                    "example": "John Doe"
                },
                "requirements": {
                    "type": "string",
                    "example": "Needs microphone"
                },
                "role": {
                    "type": "string",
                    "example": "performer"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.PublicSchedule": {
            "type": "object",
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2024-07-01T00:00:00Z"
                },
                "event_name": {
                    "type": "string",
                    "example": "Summer Music Festival"
                },
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.PublicScheduleItem"
                    }
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.PublicScheduleItem": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string",
                    "example": "Exciting show on the main stage"
                },
                "sub_items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.PublicScheduleSubItem"
                    }
                },
                "time": {
                    "type": "string",
                    "example": "2024-07-01T14:00:00Z"
                },
                "title": {
                    "type": "string",
                    "example": "Main Performance"
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.PublicScheduleSubItem": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string",
                    "example": "Opening performance"
                },
                "time": {
                    "type": "string",
                    "example": "2024-07-01T14:30:00Z"
                },
                "title": {
                    "type": "string",
                    "example": "Opening Act"
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.RiskFactor": {
            "type": "object",
            "properties": {
                "block_id": {
                    "type": "integer",
                    "example": 1
                },
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "impact": {
                    "type": "number",
                    "example": 0.7
                },
                "mitigation": {
                    "type": "string",
                    "example": "Have backup indoor venue"
                },
                "probability": {
                    "type": "number",
                    "example": 0.3
                },
                "type": {
                    "type": "string",
                    "example": "weather"
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.Schedule": {
            "type": "object",
            "required": [
                "end_date",
                "name",
                "start_date"
            ],
            "properties": {
                "blocks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Block"
                    }
                },
                "buffer_time": {
                    "type": "integer",
                    "example": 30
                },
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string",
                    "example": "Three day music festival"
                },
                "end_date": {
                    "type": "string",
                    "example": "2024-07-03T22:00:00Z"
                },
                "event_id": {
                    "type": "integer",
                    "example": 1
                },
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "name": {
                    "type": "string",
                    "example": "Summer Music Festival 2024"
                },
                "risk_score": {
                    "type": "number",
                    "example": 0.35
                },
                "start_date": {
                    "type": "string",
                    "example": "2024-07-01T10:00:00Z"
                },
                "total_duration": {
                    "type": "integer",
                    "example": 480
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.VolunteerSchedule": {
            "type": "object",
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2024-07-01T00:00:00Z"
                },
                "event_name": {
                    "type": "string",
                    "example": "Summer Music Festival"
                },
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.VolunteerScheduleItem"
                    }
                },
                "notes": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "[\"Wear volunteer badge\"",
                        "\"Follow safety guidelines\"]"
                    ]
                }
            }
        },
        "cor-events-scheduler_internal_domain_models.VolunteerScheduleItem": {
            "type": "object",
            "properties": {
                "break_duration": {
                    "type": "integer",
                    "example": 15
                },
                "equipment": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "[\"Microphones\"",
                        "\"Speakers\"]"
                    ]
                },
                "instructions": {
                    "type": "string",
                    "example": "Help with equipment setup"
                },
                "location": {
                    "type": "string",
                    "example": "Main Stage"
                },
                "required_staff": {
                    "type": "integer",
                    "example": 5
                },
                "setup_notes": {
                    "type": "string",
                    "example": "Check all connections"
                },
                "tech_break": {
                    "type": "boolean",
                    "example": false
                },
                "time": {
                    "type": "string",
                    "example": "2024-07-01T13:00:00Z"
                },
                "title": {
                    "type": "string",
                    "example": "Stage Setup"
                }
            }
        },
        "internal_handlers.AnalysisResponse": {
            "type": "object",
            "properties": {
                "optimized_schedule": {
                    "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                },
                "recommendations": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "risk_score": {
                    "type": "number"
                },
                "schedule": {
                    "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                },
                "time_analysis": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "additionalProperties": true
                    }
                }
            }
        },
        "internal_handlers.ErrorResponse": {
            "type": "object",
            "properties": {
                "details": {
                    "type": "string"
                },
                "error": {
                    "type": "string"
                }
            }
        },
        "internal_handlers.ListSchedulesResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                    }
                },
                "meta": {
                    "type": "object",
                    "properties": {
                        "page": {
                            "type": "integer"
                        },
                        "page_size": {
                            "type": "integer"
                        },
                        "total": {
                            "type": "integer"
                        }
                    }
                }
            }
        },
        "internal_handlers.OptimizationResponse": {
            "type": "object",
            "properties": {
                "improvements": {
                    "type": "object",
                    "additionalProperties": true
                },
                "optimized_schedule": {
                    "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                },
                "original_schedule": {
                    "$ref": "#/definitions/cor-events-scheduler_internal_domain_models.Schedule"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8282",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Event Scheduler API",
	Description:      "Service for managing event schedules with risk analysis and optimization",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
