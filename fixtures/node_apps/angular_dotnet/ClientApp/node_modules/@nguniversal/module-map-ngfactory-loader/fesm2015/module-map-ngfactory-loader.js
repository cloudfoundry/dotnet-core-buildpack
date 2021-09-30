import { InjectionToken, Compiler, Injectable, Inject, NgModule, NgModuleFactoryLoader } from '@angular/core';

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
/**
 * Token used by the ModuleMapNgFactoryLoader to load modules
 * @type {?}
 */
const MODULE_MAP = new InjectionToken('MODULE_MAP');
/**
 * NgModuleFactoryLoader which does not lazy load
 */
class ModuleMapNgFactoryLoader {
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

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
/**
 * Helper function for getting the providers object for the MODULE_MAP
 *
 * @param {?} moduleMap Map to use as a value for MODULE_MAP
 * @return {?}
 */
function provideModuleMap(moduleMap) {
    return {
        provide: MODULE_MAP,
        useValue: moduleMap
    };
}
/**
 * Module for using a NgModuleFactoryLoader which does not lazy load
 */
class ModuleMapLoaderModule {
    /**
     * Returns a ModuleMapLoaderModule along with a MODULE_MAP
     *
     * @param {?} moduleMap Map to use as a value for MODULE_MAP
     * @return {?}
     */
    static withMap(moduleMap) {
        return {
            ngModule: ModuleMapLoaderModule,
            providers: [
                {
                    provide: MODULE_MAP,
                    useValue: moduleMap
                }
            ]
        };
    }
}
ModuleMapLoaderModule.decorators = [
    { type: NgModule, args: [{
                providers: [
                    {
                        provide: NgModuleFactoryLoader,
                        useClass: ModuleMapNgFactoryLoader
                    }
                ]
            },] }
];

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */

/**
 * Generated bundle index. Do not edit.
 */

export { provideModuleMap, ModuleMapLoaderModule, MODULE_MAP, ModuleMapNgFactoryLoader };
//# sourceMappingURL=module-map-ngfactory-loader.js.map
