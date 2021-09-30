import { __decorate, __param, __metadata } from 'tslib';
import { InjectionToken, Compiler, Injectable, Inject, NgModule, NgModuleFactoryLoader } from '@angular/core';

/**
 * Token used by the ModuleMapNgFactoryLoader to load modules
 */
var MODULE_MAP = new InjectionToken('MODULE_MAP');
/**
 * NgModuleFactoryLoader which does not lazy load
 */
var ModuleMapNgFactoryLoader = /** @class */ (function () {
    function ModuleMapNgFactoryLoader(compiler, moduleMap) {
        this.compiler = compiler;
        this.moduleMap = moduleMap;
    }
    ModuleMapNgFactoryLoader.prototype.load = function (loadChildrenString) {
        var offlineMode = this.compiler instanceof Compiler;
        var type = this.moduleMap[loadChildrenString];
        if (!type) {
            throw new Error(loadChildrenString + " did not exist in the MODULE_MAP");
        }
        return offlineMode ?
            this.loadFactory(type) : this.loadAndCompile(type);
    };
    ModuleMapNgFactoryLoader.prototype.loadFactory = function (factory) {
        return new Promise(function (resolve) { return resolve(factory); });
    };
    ModuleMapNgFactoryLoader.prototype.loadAndCompile = function (type) {
        return this.compiler.compileModuleAsync(type);
    };
    ModuleMapNgFactoryLoader = __decorate([
        Injectable(),
        __param(1, Inject(MODULE_MAP)),
        __metadata("design:paramtypes", [Compiler, Object])
    ], ModuleMapNgFactoryLoader);
    return ModuleMapNgFactoryLoader;
}());

/**
 * Helper function for getting the providers object for the MODULE_MAP
 *
 * @param moduleMap Map to use as a value for MODULE_MAP
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
var ModuleMapLoaderModule = /** @class */ (function () {
    function ModuleMapLoaderModule() {
    }
    ModuleMapLoaderModule_1 = ModuleMapLoaderModule;
    /**
     * Returns a ModuleMapLoaderModule along with a MODULE_MAP
     *
     * @param moduleMap Map to use as a value for MODULE_MAP
     */
    ModuleMapLoaderModule.withMap = function (moduleMap) {
        return {
            ngModule: ModuleMapLoaderModule_1,
            providers: [
                {
                    provide: MODULE_MAP,
                    useValue: moduleMap
                }
            ]
        };
    };
    var ModuleMapLoaderModule_1;
    ModuleMapLoaderModule = ModuleMapLoaderModule_1 = __decorate([
        NgModule({
            providers: [
                {
                    provide: NgModuleFactoryLoader,
                    useClass: ModuleMapNgFactoryLoader
                }
            ]
        })
    ], ModuleMapLoaderModule);
    return ModuleMapLoaderModule;
}());

/**
 * @license
 * Copyright Google LLC All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */

/**
 * @license
 * Copyright Google LLC All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */

/**
 * @license
 * Copyright Google LLC All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */

/**
 * Generated bundle index. Do not edit.
 */

export { provideModuleMap, ModuleMapLoaderModule, MODULE_MAP, ModuleMapNgFactoryLoader };
//# sourceMappingURL=module-map-ngfactory-loader.js.map
