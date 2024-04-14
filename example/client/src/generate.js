import { createClient } from '@hey-api/openapi-ts'

createClient({
    input: 'http://localhost:3975/swagger-ui/swagger.json',
    output: 'src/api',
})