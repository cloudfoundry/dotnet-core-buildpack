(function (global, factory) {
    typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports, require('tslib'), require('@angular/core')) :
    typeof define === 'function' && define.amd ? define('@nguniversal/module-map-ngfactory-loader', ['exports', 'tslib', '@angular/core'], factory) :
    (global = global || self, factory((global.nguniversal = global.nguniversal || {}, global.nguniversal.moduleMapNgfactoryLoader = {}), global.tslib, global.ng.core));
}(this, function (exports, tslib_1, core) { 'use strict';

    /**
     * Token used by the ModuleMapNgFactoryLoader to load modules
     */
    var MODULE_MAP = new core.InjectionToken('MODULE_MAP');
    /**
     * NgModuleFactoryLoader which does not lazy load
     */
    var ModuleMapNgFactoryLoader = /** @class */ (function () {
        function ModuleMapNgFactoryLoader(compiler, moduleMap) {
            this.compiler = compiler;
            this.moduleMap = moduleMap;
        }
        ModuleMapNgFactoryLoader.prototype.load = function (loadChildrenString) {
            var offlineMode = this.compiler instanceof core.Compiler;
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
        ModuleMapNgFactoryLoader = tslib_1.__decorate([
            core.Injectable(),
            tslib_1.__param(1, core.Inject(MODULE_MAP)),
            tslib_1.__metadata("design:paramtypes", [core.Compiler, Object])
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
        ModuleMapLoaderModule = ModuleMapLoaderModule_1 = tslib_1.__decorate([
            core.NgModule({
                providers: [
                    {
                        provide: core.NgModuleFactoryLoader,
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

    exports.provideModuleMap = provideModuleMap;
    exports.ModuleMapLoaderModule = ModuleMapLoaderModule;
    exports.MODULE_MAP = MODULE_MAP;
    exports.ModuleMapNgFactoryLoader = ModuleMapNgFactoryLoader;

    Object.defineProperty(exports, '__esModule', { value: true });

}));
//# sourceMappingURL=module-map-ngfactory-loader.umd.js.map
