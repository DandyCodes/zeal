// This file is auto-generated by @hey-api/openapi-ts


export const $Item = {
    properties: {
        Name: {
            type: 'string'
        },
        Price: {
            type: 'number'
        }
    },
    required: ['Name', 'Price', 'Name', 'Price'],
    type: 'object'
} as const;

export const $Menu = {
    properties: {
        ID: {
            type: 'integer'
        },
        Items: {
            items: {
                '$ref': '#/components/schemas/Item'
            },
            nullable: true,
            type: 'array'
        }
    },
    required: ['ID', 'Items', 'ID', 'Items'],
    type: 'object'
} as const;