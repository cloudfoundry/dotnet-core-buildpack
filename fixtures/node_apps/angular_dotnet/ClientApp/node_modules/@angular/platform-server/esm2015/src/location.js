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
import { DOCUMENT } from '@angular/common';
import { Inject, Injectable, Optional } from '@angular/core';
import { ɵgetDOM as getDOM } from '@angular/platform-browser';
import { Subject } from 'rxjs';
import * as url from 'url';
import { INITIAL_CONFIG } from './tokens';
/**
 * @param {?} urlStr
 * @return {?}
 */
function parseUrl(urlStr) {
    /** @type {?} */
    const parsedUrl = url.parse(urlStr);
    return {
        hostname: parsedUrl.hostname || '',
        protocol: parsedUrl.protocol || '',
        port: parsedUrl.port || '',
        pathname: parsedUrl.pathname || '',
        search: parsedUrl.search || '',
        hash: parsedUrl.hash || '',
    };
}
/**
 * Server-side implementation of URL state. Implements `pathname`, `search`, and `hash`
 * but not the state stack.
 */
export class ServerPlatformLocation {
    /**
     * @param {?} _doc
     * @param {?} _config
     */
    constructor(_doc, _config) {
        this._doc = _doc;
        this.href = '/';
        this.hostname = '/';
        this.protocol = '/';
        this.port = '/';
        this.pathname = '/';
        this.search = '';
        this.hash = '';
        this._hashUpdate = new Subject();
        /** @type {?} */
        const config = (/** @type {?} */ (_config));
        if (!!config && !!config.url) {
            /** @type {?} */
            const parsedUrl = parseUrl(config.url);
            this.hostname = parsedUrl.hostname;
            this.protocol = parsedUrl.protocol;
            this.port = parsedUrl.port;
            this.pathname = parsedUrl.pathname;
            this.search = parsedUrl.search;
            this.hash = parsedUrl.hash;
        }
    }
    /**
     * @return {?}
     */
    getBaseHrefFromDOM() { return (/** @type {?} */ (getDOM().getBaseHref(this._doc))); }
    /**
     * @param {?} fn
     * @return {?}
     */
    onPopState(fn) {
        // No-op: a state stack is not implemented, so
        // no events will ever come.
    }
    /**
     * @param {?} fn
     * @return {?}
     */
    onHashChange(fn) { this._hashUpdate.subscribe(fn); }
    /**
     * @return {?}
     */
    get url() { return `${this.pathname}${this.search}${this.hash}`; }
    /**
     * @private
     * @param {?} value
     * @param {?} oldUrl
     * @return {?}
     */
    setHash(value, oldUrl) {
        if (this.hash === value) {
            // Don't fire events if the hash has not changed.
            return;
        }
        ((/** @type {?} */ (this))).hash = value;
        /** @type {?} */
        const newUrl = this.url;
        scheduleMicroTask((/**
         * @return {?}
         */
        () => this._hashUpdate.next((/** @type {?} */ ({
            type: 'hashchange', state: null, oldUrl, newUrl
        })))));
    }
    /**
     * @param {?} state
     * @param {?} title
     * @param {?} newUrl
     * @return {?}
     */
    replaceState(state, title, newUrl) {
        /** @type {?} */
        const oldUrl = this.url;
        /** @type {?} */
        const parsedUrl = parseUrl(newUrl);
        ((/** @type {?} */ (this))).pathname = parsedUrl.pathname;
        ((/** @type {?} */ (this))).search = parsedUrl.search;
        this.setHash(parsedUrl.hash, oldUrl);
    }
    /**
     * @param {?} state
     * @param {?} title
     * @param {?} newUrl
     * @return {?}
     */
    pushState(state, title, newUrl) {
        this.replaceState(state, title, newUrl);
    }
    /**
     * @return {?}
     */
    forward() { throw new Error('Not implemented'); }
    /**
     * @return {?}
     */
    back() { throw new Error('Not implemented'); }
    // History API isn't available on server, therefore return undefined
    /**
     * @return {?}
     */
    getState() { return undefined; }
}
ServerPlatformLocation.decorators = [
    { type: Injectable }
];
/** @nocollapse */
ServerPlatformLocation.ctorParameters = () => [
    { type: undefined, decorators: [{ type: Inject, args: [DOCUMENT,] }] },
    { type: undefined, decorators: [{ type: Optional }, { type: Inject, args: [INITIAL_CONFIG,] }] }
];
if (false) {
    /** @type {?} */
    ServerPlatformLocation.prototype.href;
    /** @type {?} */
    ServerPlatformLocation.prototype.hostname;
    /** @type {?} */
    ServerPlatformLocation.prototype.protocol;
    /** @type {?} */
    ServerPlatformLocation.prototype.port;
    /** @type {?} */
    ServerPlatformLocation.prototype.pathname;
    /** @type {?} */
    ServerPlatformLocation.prototype.search;
    /** @type {?} */
    ServerPlatformLocation.prototype.hash;
    /**
     * @type {?}
     * @private
     */
    ServerPlatformLocation.prototype._hashUpdate;
    /**
     * @type {?}
     * @private
     */
    ServerPlatformLocation.prototype._doc;
}
/**
 * @param {?} fn
 * @return {?}
 */
export function scheduleMicroTask(fn) {
    Zone.current.scheduleMicroTask('scheduleMicrotask', fn);
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibG9jYXRpb24uanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi8uLi8uLi8uLi8uLi8uLi9wYWNrYWdlcy9wbGF0Zm9ybS1zZXJ2ZXIvc3JjL2xvY2F0aW9uLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiI7Ozs7Ozs7Ozs7O0FBUUEsT0FBTyxFQUFDLFFBQVEsRUFBZ0UsTUFBTSxpQkFBaUIsQ0FBQztBQUN4RyxPQUFPLEVBQUMsTUFBTSxFQUFFLFVBQVUsRUFBRSxRQUFRLEVBQUMsTUFBTSxlQUFlLENBQUM7QUFDM0QsT0FBTyxFQUFDLE9BQU8sSUFBSSxNQUFNLEVBQUMsTUFBTSwyQkFBMkIsQ0FBQztBQUM1RCxPQUFPLEVBQUMsT0FBTyxFQUFDLE1BQU0sTUFBTSxDQUFDO0FBQzdCLE9BQU8sS0FBSyxHQUFHLE1BQU0sS0FBSyxDQUFDO0FBQzNCLE9BQU8sRUFBQyxjQUFjLEVBQWlCLE1BQU0sVUFBVSxDQUFDOzs7OztBQUd4RCxTQUFTLFFBQVEsQ0FBQyxNQUFjOztVQUN4QixTQUFTLEdBQUcsR0FBRyxDQUFDLEtBQUssQ0FBQyxNQUFNLENBQUM7SUFDbkMsT0FBTztRQUNMLFFBQVEsRUFBRSxTQUFTLENBQUMsUUFBUSxJQUFJLEVBQUU7UUFDbEMsUUFBUSxFQUFFLFNBQVMsQ0FBQyxRQUFRLElBQUksRUFBRTtRQUNsQyxJQUFJLEVBQUUsU0FBUyxDQUFDLElBQUksSUFBSSxFQUFFO1FBQzFCLFFBQVEsRUFBRSxTQUFTLENBQUMsUUFBUSxJQUFJLEVBQUU7UUFDbEMsTUFBTSxFQUFFLFNBQVMsQ0FBQyxNQUFNLElBQUksRUFBRTtRQUM5QixJQUFJLEVBQUUsU0FBUyxDQUFDLElBQUksSUFBSSxFQUFFO0tBQzNCLENBQUM7QUFDSixDQUFDOzs7OztBQU9ELE1BQU0sT0FBTyxzQkFBc0I7Ozs7O0lBVWpDLFlBQzhCLElBQVMsRUFBc0MsT0FBWTtRQUEzRCxTQUFJLEdBQUosSUFBSSxDQUFLO1FBVnZCLFNBQUksR0FBVyxHQUFHLENBQUM7UUFD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QUFXLE1BQU0sSUFBSSxLQUFLLENBQUMsaUJBQWlCLENBQUMsQ0FBQyxDQUFDLENBQUM7Ozs7SUFFdkQsSUFBSSxLQUFXLE1BQU0sSUFBSSxLQUFLLENBQUMsaUJBQWlCLENBQUMsQ0FBQyxDQUFDLENBQUM7Ozs7O0lBR3BELFFBQVEsS0FBYyxPQUFPLFNBQVMsQ0FBQyxDQUFDLENBQUM7OztZQWpFMUMsVUFBVTs7Ozs0Q0FZSixNQUFNLFNBQUMsUUFBUTs0Q0FBc0IsUUFBUSxZQUFJLE1BQU0sU0FBQyxjQUFjOzs7O0lBVjNFLHNDQUFtQzs7SUFDbkMsMENBQXVDOztJQUN2QywwQ0FBdUM7O0lBQ3ZDLHNDQUFtQzs7SUFDbkMsMENBQXVDOztJQUN2Qyx3Q0FBb0M7O0lBQ3BDLHNDQUFrQzs7Ozs7SUFDbEMsNkNBQXlEOzs7OztJQUdyRCxzQ0FBbUM7Ozs7OztBQXdEekMsTUFBTSxVQUFVLGlCQUFpQixDQUFDLEVBQVk7SUFDNUMsSUFBSSxDQUFDLE9BQU8sQ0FBQyxpQkFBaUIsQ0FBQyxtQkFBbUIsRUFBRSxFQUFFLENBQUMsQ0FBQztBQUMxRCxDQUFDIiwic291cmNlc0NvbnRlbnQiOlsiLyoqXG4gKiBAbGljZW5zZVxuICogQ29weXJpZ2h0IEdvb2dsZSBJbmMuIEFsbCBSaWdodHMgUmVzZXJ2ZWQuXG4gKlxuICogVXNlIG9mIHRoaXMgc291cmNlIGNvZGUgaXMgZ292ZXJuZWQgYnkgYW4gTUlULXN0eWxlIGxpY2Vuc2UgdGhhdCBjYW4gYmVcbiAqIGZvdW5kIGluIHRoZSBMSUNFTlNFIGZpbGUgYXQgaHR0cHM6Ly9hbmd1bGFyLmlvL2xpY2Vuc2VcbiAqL1xuXG5pbXBvcnQge0RPQ1VNRU5ULCBMb2NhdGlvbkNoYW5nZUV2ZW50LCBMb2NhdGlvbkNoYW5nZUxpc3RlbmVyLCBQbGF0Zm9ybUxvY2F0aW9ufSBmcm9tICdAYW5ndWxhci9jb21tb24nO1xuaW1wb3J0IHtJbmplY3QsIEluamVjdGFibGUsIE9wdGlvbmFsfSBmcm9tICdAYW5ndWxhci9jb3JlJztcbmltcG9ydCB7ybVnZXRET00gYXMgZ2V0RE9NfSBmcm9tICdAYW5ndWxhci9wbGF0Zm9ybS1icm93c2VyJztcbmltcG9ydCB7U3ViamVjdH0gZnJvbSAncnhqcyc7XG5pbXBvcnQgKiBhcyB1cmwgZnJvbSAndXJsJztcbmltcG9ydCB7SU5JVElBTF9DT05GSUcsIFBsYXRmb3JtQ29uZmlnfSBmcm9tICcuL3Rva2Vucyc7XG5cblxuZnVuY3Rpb24gcGFyc2VVcmwodXJsU3RyOiBzdHJpbmcpIHtcbiAgY29uc3QgcGFyc2VkVXJsID0gdXJsLnBhcnNlKHVybFN0cik7XG4gIHJldHVybiB7XG4gICAgaG9zdG5hbWU6IHBhcnNlZFVybC5ob3N0bmFtZSB8fCAnJyxcbiAgICBwcm90b2NvbDogcGFyc2VkVXJsLnByb3RvY29sIHx8ICcnLFxuICAgIHBvcnQ6IHBhcnNlZFVybC5wb3J0IHx8ICcnLFxuICAgIHBhdGhuYW1lOiBwYXJzZWRVcmwucGF0aG5hbWUgfHwgJycsXG4gICAgc2VhcmNoOiBwYXJzZWRVcmwuc2VhcmNoIHx8ICcnLFxuICAgIGhhc2g6IHBhcnNlZFVybC5oYXNoIHx8ICcnLFxuICB9O1xufVxuXG4vKipcbiAqIFNlcnZlci1zaWRlIGltcGxlbWVudGF0aW9uIG9mIFVSTCBzdGF0ZS4gSW1wbGVtZW50cyBgcGF0aG5hbWVgLCBgc2VhcmNoYCwgYW5kIGBoYXNoYFxuICogYnV0IG5vdCB0aGUgc3RhdGUgc3RhY2suXG4gKi9cbkBJbmplY3RhYmxlKClcbmV4cG9ydCBjbGFzcyBTZXJ2ZXJQbGF0Zm9ybUxvY2F0aW9uIGltcGxlbWVudHMgUGxhdGZvcm1Mb2NhdGlvbiB7XG4gIHB1YmxpYyByZWFkb25seSBocmVmOiBzdHJpbmcgPSAnLyc7XG4gIHB1YmxpYyByZWFkb25seSBob3N0bmFtZTogc3RyaW5nID0gJy8nO1xuICBwdWJsaWMgcmVhZG9ubHkgcHJvdG9jb2w6IHN0cmluZyA9ICcvJztcbiAgcHVibGljIHJlYWRvbmx5IHBvcnQ6IHN0cmluZyA9ICcvJztcbiAgcHVibGljIHJlYWRvbmx5IHBhdGhuYW1lOiBzdHJpbmcgPSAnLyc7XG4gIHB1YmxpYyByZWFkb25seSBzZWFyY2g6IHN0cmluZyA9ICcnO1xuICBwdWJsaWMgcmVhZG9ubHkgaGFzaDogc3RyaW5nID0gJyc7XG4gIHByaXZhdGUgX2hhc2hVcGRhdGUgPSBuZXcgU3ViamVjdDxMb2NhdGlvbkNoYW5nZUV2ZW50PigpO1xuXG4gIGNvbnN0cnVjdG9yKFxuICAgICAgQEluamVjdChET0NVTUVOVCkgcHJpdmF0ZSBfZG9jOiBhbnksIEBPcHRpb25hbCgpIEBJbmplY3QoSU5JVElBTF9DT05GSUcpIF9jb25maWc6IGFueSkge1xuICAgIGNvbnN0IGNvbmZpZyA9IF9jb25maWcgYXMgUGxhdGZvcm1Db25maWcgfCBudWxsO1xuICAgIGlmICghIWNvbmZpZyAmJiAhIWNvbmZpZy51cmwpIHtcbiAgICAgIGNvbnN0IHBhcnNlZFVybCA9IHBhcnNlVXJsKGNvbmZpZy51cmwpO1xuICAgICAgdGhpcy5ob3N0bmFtZSA9IHBhcnNlZFVybC5ob3N0bmFtZTtcbiAgICAgIHRoaXMucHJvdG9jb2wgPSBwYXJzZWRVcmwucHJvdG9jb2w7XG4gICAgICB0aGlzLnBvcnQgPSBwYXJzZWRVcmwucG9ydDtcbiAgICAgIHRoaXMucGF0aG5hbWUgPSBwYXJzZWRVcmwucGF0aG5hbWU7XG4gICAgICB0aGlzLnNlYXJjaCA9IHBhcnNlZFVybC5zZWFyY2g7XG4gICAgICB0aGlzLmhhc2ggPSBwYXJzZWRVcmwuaGFzaDtcbiAgICB9XG4gIH1cblxuICBnZXRCYXNlSHJlZkZyb21ET00oKTogc3RyaW5nIHsgcmV0dXJuIGdldERPTSgpLmdldEJhc2VIcmVmKHRoaXMuX2RvYykgITsgfVxuXG4gIG9uUG9wU3RhdGUoZm46IExvY2F0aW9uQ2hhbmdlTGlzdGVuZXIpOiB2b2lkIHtcbiAgICAvLyBOby1vcDogYSBzdGF0ZSBzdGFjayBpcyBub3QgaW1wbGVtZW50ZWQsIHNvXG4gICAgLy8gbm8gZXZlbnRzIHdpbGwgZXZlciBjb21lLlxuICB9XG5cbiAgb25IYXNoQ2hhbmdlKGZuOiBMb2NhdGlvbkNoYW5nZUxpc3RlbmVyKTogdm9pZCB7IHRoaXMuX2hhc2hVcGRhdGUuc3Vic2NyaWJlKGZuKTsgfVxuXG4gIGdldCB1cmwoKTogc3RyaW5nIHsgcmV0dXJuIGAke3RoaXMucGF0aG5hbWV9JHt0aGlzLnNlYXJjaH0ke3RoaXMuaGFzaH1gOyB9XG5cbiAgcHJpdmF0ZSBzZXRIYXNoKHZhbHVlOiBzdHJpbmcsIG9sZFVybDogc3RyaW5nKSB7XG4gICAgaWYgKHRoaXMuaGFzaCA9PT0gdmFsdWUpIHtcbiAgICAgIC8vIERvbid0IGZpcmUgZXZlbnRzIGlmIHRoZSBoYXNoIGhhcyBub3QgY2hhbmdlZC5cbiAgICAgIHJldHVybjtcbiAgICB9XG4gICAgKHRoaXMgYXN7aGFzaDogc3RyaW5nfSkuaGFzaCA9IHZhbHVlO1xuICAgIGNvbnN0IG5ld1VybCA9IHRoaXMudXJsO1xuICAgIHNjaGVkdWxlTWljcm9UYXNrKCgpID0+IHRoaXMuX2hhc2hVcGRhdGUubmV4dCh7XG4gICAgICB0eXBlOiAnaGFzaGNoYW5nZScsIHN0YXRlOiBudWxsLCBvbGRVcmwsIG5ld1VybFxuICAgIH0gYXMgTG9jYXRpb25DaGFuZ2VFdmVudCkpO1xuICB9XG5cbiAgcmVwbGFjZVN0YXRlKHN0YXRlOiBhbnksIHRpdGxlOiBzdHJpbmcsIG5ld1VybDogc3RyaW5nKTogdm9pZCB7XG4gICAgY29uc3Qgb2xkVXJsID0gdGhpcy51cmw7XG4gICAgY29uc3QgcGFyc2VkVXJsID0gcGFyc2VVcmwobmV3VXJsKTtcbiAgICAodGhpcyBhc3twYXRobmFtZTogc3RyaW5nfSkucGF0aG5hbWUgPSBwYXJzZWRVcmwucGF0aG5hbWU7XG4gICAgKHRoaXMgYXN7c2VhcmNoOiBzdHJpbmd9KS5zZWFyY2ggPSBwYXJzZWRVcmwuc2VhcmNoO1xuICAgIHRoaXMuc2V0SGFzaChwYXJzZWRVcmwuaGFzaCwgb2xkVXJsKTtcbiAgfVxuXG4gIHB1c2hTdGF0ZShzdGF0ZTogYW55LCB0aXRsZTogc3RyaW5nLCBuZXdVcmw6IHN0cmluZyk6IHZvaWQge1xuICAgIHRoaXMucmVwbGFjZVN0YXRlKHN0YXRlLCB0aXRsZSwgbmV3VXJsKTtcbiAgfVxuXG4gIGZvcndhcmQoKTogdm9pZCB7IHRocm93IG5ldyBFcnJvcignTm90IGltcGxlbWVudGVkJyk7IH1cblxuICBiYWNrKCk6IHZvaWQgeyB0aHJvdyBuZXcgRXJyb3IoJ05vdCBpbXBsZW1lbnRlZCcpOyB9XG5cbiAgLy8gSGlzdG9yeSBBUEkgaXNuJ3QgYXZhaWxhYmxlIG9uIHNlcnZlciwgdGhlcmVmb3JlIHJldHVybiB1bmRlZmluZWRcbiAgZ2V0U3RhdGUoKTogdW5rbm93biB7IHJldHVybiB1bmRlZmluZWQ7IH1cbn1cblxuZXhwb3J0IGZ1bmN0aW9uIHNjaGVkdWxlTWljcm9UYXNrKGZuOiBGdW5jdGlvbikge1xuICBab25lLmN1cnJlbnQuc2NoZWR1bGVNaWNyb1Rhc2soJ3NjaGVkdWxlTWljcm90YXNrJywgZm4pO1xufVxuIl19