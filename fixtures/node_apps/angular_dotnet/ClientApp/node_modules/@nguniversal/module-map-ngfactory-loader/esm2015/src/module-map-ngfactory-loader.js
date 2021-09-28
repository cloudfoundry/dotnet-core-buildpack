/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
/**
 * @license
 * Copyright Google LLC All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { Injectable, InjectionToken, Inject, Compiler } from '@angular/core';
/**
 * Token used by the ModuleMapNgFactoryLoader to load modules
 * @type {?}
 */
export const MODULE_MAP = new InjectionToken('MODULE_MAP');
/**
 * NgModuleFactoryLoader which does not lazy load
 */
export class ModuleMapNgFactoryLoader {
    /**
     * @param {?} compiler
     * @param {?} moduleMap
     */
    constructor(compiler, moduleMap) {
        this.compiler = compiler;
        this.moduleMap = moduleMap;
    }
    /**
     * @param {?} loadChildrenString
     * @return {?}
     */
    load(loadChildrenString) {
        /** @type {?} */
        const offlineMode = this.compiler instanceof Compiler;
        /** @type {?} */
        const type = this.moduleMap[loadChildrenString];
        if (!type) {
            throw new Error(`${loadChildrenString} did not exist in the MODULE_MAP`);
        }
        return offlineMode ?
            this.loadFactory((/** @type {?} */ (type))) : this.loadAndCompile((/** @type {?} */ (type)));
    }
    /**
     * @private
     * @param {?} factory
     * @return {?}
     */
    loadFactory(factory) {
        return new Promise((/**
         * @param {?} resolve
         * @return {?}
         */
        resolve => resolve(factory)));
    }
    /**
     * @private
     * @param {?} type
     * @return {?}
     */
    loadAndCompile(type) {
        return this.compiler.compileModuleAsync(type);
    }
}
ModuleMapNgFactoryLoader.decorators = [
    { type: Injectable }
];
/** @nocollapse */
ModuleMapNgFactoryLoader.ctorParameters = () => [
    { type: Compiler },
    { type: undefined, decorators: [{ type: Inject, args: [MODULE_MAP,] }] }
];
if (false) {
    /**
     * @type {?}
     * @private
     */
    ModuleMapNgFactoryLoader.prototype.compiler;
    /**
     * @type {?}
     * @private
     */
    ModuleMapNgFactoryLoader.prototype.moduleMap;
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibW9kdWxlLW1hcC1uZ2ZhY3RvcnktbG9hZGVyLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiLi4vLi4vLi4vLi4vLi4vLi4vbW9kdWxlcy9tb2R1bGUtbWFwLW5nZmFjdG9yeS1sb2FkZXIvc3JjL21vZHVsZS1tYXAtbmdmYWN0b3J5LWxvYWRlci50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7Ozs7Ozs7OztBQU9BLE9BQU8sRUFDTCxVQUFVLEVBRVYsY0FBYyxFQUVkLE1BQU0sRUFFTixRQUFRLEVBQ1QsTUFBTSxlQUFlLENBQUM7Ozs7O0FBTXZCLE1BQU0sT0FBTyxVQUFVLEdBQThCLElBQUksY0FBYyxDQUFDLFlBQVksQ0FBQzs7OztBQU1yRixNQUFNLE9BQU8sd0JBQXdCOzs7OztJQUNuQyxZQUFvQixRQUFrQixFQUE4QixTQUFvQjtRQUFwRSxhQUFRLEdBQVIsUUFBUSxDQUFVO1FBQThCLGNBQVMsR0FBVCxTQUFTLENBQVc7SUFBSSxDQUFDOzs7OztJQUU3RixJQUFJLENBQUMsa0JBQTBCOztjQUN2QixXQUFXLEdBQUcsSUFBSSxDQUFDLFFBQVEsWUFBWSxRQUFROztjQUMvQyxJQUFJLEdBQUcsSUFBSSxDQUFDLFNBQVMsQ0FBQyxrQkFBa0IsQ0FBQztRQUUvQyxJQUFJLENBQUMsSUFBSSxFQUFFO1lBQ1QsTUFBTSxJQUFJLEtBQUssQ0FBQyxHQUFHLGtCQUFrQixrQ0FBa0MsQ0FBQyxDQUFDO1NBQzFFO1FBRUQsT0FBTyxXQUFXLENBQUMsQ0FBQztZQUNsQixJQUFJLENBQUMsV0FBVyxDQUFDLG1CQUF1QixJQUFJLEVBQUEsQ0FBQyxDQUFDLENBQUMsQ0FBQyxJQUFJLENBQUMsY0FBYyxDQUFDLG1CQUFZLElBQUksRUFBQSxDQUFDLENBQUM7SUFDMUYsQ0FBQzs7Ozs7O0lBRU8sV0FBVyxDQUFDLE9BQTZCO1FBQy9DLE9BQU8sSUFBSSxPQUFPOzs7O1FBQUMsT0FBTyxDQUFDLEVBQUUsQ0FBQyxPQUFPLENBQUMsT0FBTyxDQUFDLEVBQUMsQ0FBQztJQUNsRCxDQUFDOzs7Ozs7SUFFTyxjQUFjLENBQUMsSUFBZTtRQUNwQyxPQUFPLElBQUksQ0FBQyxRQUFRLENBQUMsa0JBQWtCLENBQUMsSUFBSSxDQUFDLENBQUM7SUFDaEQsQ0FBQzs7O1lBdEJGLFVBQVU7Ozs7WUFaVCxRQUFROzRDQWNpQyxNQUFNLFNBQUMsVUFBVTs7Ozs7OztJQUE5Qyw0Q0FBMEI7Ozs7O0lBQUUsNkNBQWdEIiwic291cmNlc0NvbnRlbnQiOlsiLyoqXG4gKiBAbGljZW5zZVxuICogQ29weXJpZ2h0IEdvb2dsZSBMTEMgQWxsIFJpZ2h0cyBSZXNlcnZlZC5cbiAqXG4gKiBVc2Ugb2YgdGhpcyBzb3VyY2UgY29kZSBpcyBnb3Zlcm5lZCBieSBhbiBNSVQtc3R5bGUgbGljZW5zZSB0aGF0IGNhbiBiZVxuICogZm91bmQgaW4gdGhlIExJQ0VOU0UgZmlsZSBhdCBodHRwczovL2FuZ3VsYXIuaW8vbGljZW5zZVxuICovXG5pbXBvcnQge1xuICBJbmplY3RhYmxlLFxuICBOZ01vZHVsZUZhY3RvcnlMb2FkZXIsXG4gIEluamVjdGlvblRva2VuLFxuICBOZ01vZHVsZUZhY3RvcnksXG4gIEluamVjdCxcbiAgVHlwZSxcbiAgQ29tcGlsZXJcbn0gZnJvbSAnQGFuZ3VsYXIvY29yZSc7XG5pbXBvcnQge01vZHVsZU1hcH0gZnJvbSAnLi9tb2R1bGUtbWFwJztcblxuLyoqXG4gKiBUb2tlbiB1c2VkIGJ5IHRoZSBNb2R1bGVNYXBOZ0ZhY3RvcnlMb2FkZXIgdG8gbG9hZCBtb2R1bGVzXG4gKi9cbmV4cG9ydCBjb25zdCBNT0RVTEVfTUFQOiBJbmplY3Rpb25Ub2tlbjxNb2R1bGVNYXA+ID0gbmV3IEluamVjdGlvblRva2VuKCdNT0RVTEVfTUFQJyk7XG5cbi8qKlxuICogTmdNb2R1bGVGYWN0b3J5TG9hZGVyIHdoaWNoIGRvZXMgbm90IGxhenkgbG9hZFxuICovXG5ASW5qZWN0YWJsZSgpXG5leHBvcnQgY2xhc3MgTW9kdWxlTWFwTmdGYWN0b3J5TG9hZGVyIGltcGxlbWVudHMgTmdNb2R1bGVGYWN0b3J5TG9hZGVyIHtcbiAgY29uc3RydWN0b3IocHJpdmF0ZSBjb21waWxlcjogQ29tcGlsZXIsIEBJbmplY3QoTU9EVUxFX01BUCkgcHJpdmF0ZSBtb2R1bGVNYXA6IE1vZHVsZU1hcCkgeyB9XG5cbiAgbG9hZChsb2FkQ2hpbGRyZW5TdHJpbmc6IHN0cmluZyk6IFByb21pc2U8TmdNb2R1bGVGYWN0b3J5PGFueT4+IHtcbiAgICBjb25zdCBvZmZsaW5lTW9kZSA9IHRoaXMuY29tcGlsZXIgaW5zdGFuY2VvZiBDb21waWxlcjtcbiAgICBjb25zdCB0eXBlID0gdGhpcy5tb2R1bGVNYXBbbG9hZENoaWxkcmVuU3RyaW5nXTtcblxuICAgIGlmICghdHlwZSkge1xuICAgICAgdGhyb3cgbmV3IEVycm9yKGAke2xvYWRDaGlsZHJlblN0cmluZ30gZGlkIG5vdCBleGlzdCBpbiB0aGUgTU9EVUxFX01BUGApO1xuICAgIH1cblxuICAgIHJldHVybiBvZmZsaW5lTW9kZSA/XG4gICAgICB0aGlzLmxvYWRGYWN0b3J5KDxOZ01vZHVsZUZhY3Rvcnk8YW55Pj4gdHlwZSkgOiB0aGlzLmxvYWRBbmRDb21waWxlKDxUeXBlPGFueT4+IHR5cGUpO1xuICB9XG5cbiAgcHJpdmF0ZSBsb2FkRmFjdG9yeShmYWN0b3J5OiBOZ01vZHVsZUZhY3Rvcnk8YW55Pik6IFByb21pc2U8TmdNb2R1bGVGYWN0b3J5PGFueT4+IHtcbiAgICByZXR1cm4gbmV3IFByb21pc2UocmVzb2x2ZSA9PiByZXNvbHZlKGZhY3RvcnkpKTtcbiAgfVxuXG4gIHByaXZhdGUgbG9hZEFuZENvbXBpbGUodHlwZTogVHlwZTxhbnk+KTogUHJvbWlzZTxOZ01vZHVsZUZhY3Rvcnk8YW55Pj4ge1xuICAgIHJldHVybiB0aGlzLmNvbXBpbGVyLmNvbXBpbGVNb2R1bGVBc3luYyh0eXBlKTtcbiAgfVxufVxuIl19