// This file is auto-generated by @hey-api/openapi-ts

import type { CancelablePromise } from './core/CancelablePromise';
import { OpenAPI } from './core/OpenAPI';
import { request as __request } from './core/request';
import type { $OpenApiTs } from './models';

export class DefaultService {
	/**
	 * @returns number 
	 * @throws ApiError
	 */
	public static getAnswer(): CancelablePromise<$OpenApiTs['/answer']['get']['res'][200]> {
		
		return __request(OpenAPI, {
			method: 'GET',
			url: '/answer',
		});
	}

	/**
	 * @returns string 
	 * @throws ApiError
	 */
	public static postHello(): CancelablePromise<$OpenApiTs['/hello']['post']['res'][200]> {
		
		return __request(OpenAPI, {
			method: 'POST',
			url: '/hello',
		});
	}

	/**
	 * @returns string 
	 * @throws ApiError
	 */
	public static putItems(data: $OpenApiTs['/items']['put']['req']): CancelablePromise<$OpenApiTs['/items']['put']['res'][200]> {
		const {
                    requestBody
                } = data;
		return __request(OpenAPI, {
			method: 'PUT',
			url: '/items',
			body: requestBody,
			mediaType: 'application/json',
		});
	}

	/**
	 * @returns Item 
	 * @throws ApiError
	 */
	public static postItemsByMenuId(data: $OpenApiTs['/items/{MenuID}']['post']['req']): CancelablePromise<$OpenApiTs['/items/{MenuID}']['post']['res'][200]> {
		const {
                    menuId,
requestBody
                } = data;
		return __request(OpenAPI, {
			method: 'POST',
			url: '/items/{MenuID}',
			path: {
				MenuID: menuId
			},
			body: requestBody,
			mediaType: 'application/json',
		});
	}

	/**
	 * @returns Menu 
	 * @throws ApiError
	 */
	public static getMenus(): CancelablePromise<$OpenApiTs['/menus']['get']['res'][200]> {
		
		return __request(OpenAPI, {
			method: 'GET',
			url: '/menus',
		});
	}

	/**
	 * @returns string 
	 * @throws ApiError
	 */
	public static deleteMenusById(data: $OpenApiTs['/menus/{ID}']['delete']['req']): CancelablePromise<$OpenApiTs['/menus/{ID}']['delete']['res'][200]> {
		const {
                    quiet,
id
                } = data;
		return __request(OpenAPI, {
			method: 'DELETE',
			url: '/menus/{ID}',
			path: {
				ID: id
			},
			query: {
				Quiet: quiet
			},
		});
	}

}