/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,constantProperty,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { InjectionToken } from '@angular/core';
/**
 * Config object passed to initialize the platform.
 *
 * \@publicApi
 * @record
 */
export function PlatformConfig() { }
if (false) {
    /** @type {?|undefined} */
    PlatformConfig.prototype.document;
    /** @type {?|undefined} */
    PlatformConfig.prototype.url;
}
/**
 * The DI token for setting the initial config for the platform.
 *
 * \@publicApi
 * @type {?}
 */
export const INITIAL_CONFIG = new InjectionToken('Server.INITIAL_CONFIG');
/**
 * A function that will be executed when calling `renderModuleFactory` or `renderModule` just
 * before current platform state is rendered to string.
 *
 * \@publicApi
 * @type {?}
 */
export const BEFORE_APP_SERIALIZED = new InjectionToken('Server.RENDER_MODULE_HOOK');
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoidG9rZW5zLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiLi4vLi4vLi4vLi4vLi4vLi4vcGFja2FnZXMvcGxhdGZvcm0tc2VydmVyL3NyYy90b2tlbnMudHMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6Ijs7Ozs7Ozs7Ozs7QUFRQSxPQUFPLEVBQUMsY0FBYyxFQUFDLE1BQU0sZUFBZSxDQUFDOzs7Ozs7O0FBTzdDLG9DQUdDOzs7SUFGQyxrQ0FBa0I7O0lBQ2xCLDZCQUFhOzs7Ozs7OztBQVFmLE1BQU0sT0FBTyxjQUFjLEdBQUcsSUFBSSxjQUFjLENBQWlCLHVCQUF1QixDQUFDOzs7Ozs7OztBQVF6RixNQUFNLE9BQU8scUJBQXFCLEdBQzlCLElBQUksY0FBYyxDQUFvQiwyQkFBMkIsQ0FBQyIsInNvdXJjZXNDb250ZW50IjpbIi8qKlxuICogQGxpY2Vuc2VcbiAqIENvcHlyaWdodCBHb29nbGUgSW5jLiBBbGwgUmlnaHRzIFJlc2VydmVkLlxuICpcbiAqIFVzZSBvZiB0aGlzIHNvdXJjZSBjb2RlIGlzIGdvdmVybmVkIGJ5IGFuIE1JVC1zdHlsZSBsaWNlbnNlIHRoYXQgY2FuIGJlXG4gKiBmb3VuZCBpbiB0aGUgTElDRU5TRSBmaWxlIGF0IGh0dHBzOi8vYW5ndWxhci5pby9saWNlbnNlXG4gKi9cblxuaW1wb3J0IHtJbmplY3Rpb25Ub2tlbn0gZnJvbSAnQGFuZ3VsYXIvY29yZSc7XG5cbi8qKlxuICogQ29uZmlnIG9iamVjdCBwYXNzZWQgdG8gaW5pdGlhbGl6ZSB0aGUgcGxhdGZvcm0uXG4gKlxuICogQHB1YmxpY0FwaVxuICovXG5leHBvcnQgaW50ZXJmYWNlIFBsYXRmb3JtQ29uZmlnIHtcbiAgZG9jdW1lbnQ/OiBzdHJpbmc7XG4gIHVybD86IHN0cmluZztcbn1cblxuLyoqXG4gKiBUaGUgREkgdG9rZW4gZm9yIHNldHRpbmcgdGhlIGluaXRpYWwgY29uZmlnIGZvciB0aGUgcGxhdGZvcm0uXG4gKlxuICogQHB1YmxpY0FwaVxuICovXG5leHBvcnQgY29uc3QgSU5JVElBTF9DT05GSUcgPSBuZXcgSW5qZWN0aW9uVG9rZW48UGxhdGZvcm1Db25maWc+KCdTZXJ2ZXIuSU5JVElBTF9DT05GSUcnKTtcblxuLyoqXG4gKiBBIGZ1bmN0aW9uIHRoYXQgd2lsbCBiZSBleGVjdXRlZCB3aGVuIGNhbGxpbmcgYHJlbmRlck1vZHVsZUZhY3RvcnlgIG9yIGByZW5kZXJNb2R1bGVgIGp1c3RcbiAqIGJlZm9yZSBjdXJyZW50IHBsYXRmb3JtIHN0YXRlIGlzIHJlbmRlcmVkIHRvIHN0cmluZy5cbiAqXG4gKiBAcHVibGljQXBpXG4gKi9cbmV4cG9ydCBjb25zdCBCRUZPUkVfQVBQX1NFUklBTElaRUQgPVxuICAgIG5ldyBJbmplY3Rpb25Ub2tlbjxBcnJheTwoKSA9PiB2b2lkPj4oJ1NlcnZlci5SRU5ERVJfTU9EVUxFX0hPT0snKTtcbiJdfQ==