cd src && swag init && docker run -p 80:8080 -e SWAGGER_JSON=/swagger.json -v `pwd`/docs/swagger.json:/swagger.json swaggerapi/swagger-ui
