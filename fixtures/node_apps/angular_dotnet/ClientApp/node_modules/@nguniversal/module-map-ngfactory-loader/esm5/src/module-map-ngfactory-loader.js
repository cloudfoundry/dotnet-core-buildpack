import * as tslib_1 from "tslib";
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
 */
export var MODULE_MAP = new InjectionToken('MODULE_MAP');
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
    ModuleMapNgFactoryLoader = tslib_1.__decorate([
        Injectable(),
        tslib_1.__param(1, Inject(MODULE_MAP)),
        tslib_1.__metadata("design:paramtypes", [Compiler, Object])
    ], ModuleMapNgFactoryLoader);
    return ModuleMapNgFactoryLoader;
}());
export { ModuleMapNgFactoryLoader };
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibW9kdWxlLW1hcC1uZ2ZhY3RvcnktbG9hZGVyLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiLi4vLi4vLi4vLi4vLi4vLi4vLi4vLi4vLi4vbW9kdWxlcy9tb2R1bGUtbWFwLW5nZmFjdG9yeS1sb2FkZXIvc3JjL21vZHVsZS1tYXAtbmdmYWN0b3J5LWxvYWRlci50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiO0FBQUE7Ozs7OztHQU1HO0FBQ0gsT0FBTyxFQUNMLFVBQVUsRUFFVixjQUFjLEVBRWQsTUFBTSxFQUVOLFFBQVEsRUFDVCxNQUFNLGVBQWUsQ0FBQztBQUd2Qjs7R0FFRztBQUNILE1BQU0sQ0FBQyxJQUFNLFVBQVUsR0FBOEIsSUFBSSxjQUFjLENBQUMsWUFBWSxDQUFDLENBQUM7QUFFdEY7O0dBRUc7QUFFSDtJQUNFLGtDQUFvQixRQUFrQixFQUE4QixTQUFvQjtRQUFwRSxhQUFRLEdBQVIsUUFBUSxDQUFVO1FBQThCLGNBQVMsR0FBVCxTQUFTLENBQVc7SUFBSSxDQUFDO0lBRTdGLHVDQUFJLEdBQUosVUFBSyxrQkFBMEI7UUFDN0IsSUFBTSxXQUFXLEdBQUcsSUFBSSxDQUFDLFFBQVEsWUFBWSxRQUFRLENBQUM7UUFDdEQsSUFBTSxJQUFJLEdBQUcsSUFBSSxDQUFDLFNBQVMsQ0FBQyxrQkFBa0IsQ0FBQyxDQUFDO1FBRWhELElBQUksQ0FBQyxJQUFJLEVBQUU7WUFDVCxNQUFNLElBQUksS0FBSyxDQUFJLGtCQUFrQixxQ0FBa0MsQ0FBQyxDQUFDO1NBQzFFO1FBRUQsT0FBTyxXQUFXLENBQUMsQ0FBQztZQUNsQixJQUFJLENBQUMsV0FBVyxDQUF3QixJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLGNBQWMsQ0FBYSxJQUFJLENBQUMsQ0FBQztJQUMxRixDQUFDO0lBRU8sOENBQVcsR0FBbkIsVUFBb0IsT0FBNkI7UUFDL0MsT0FBTyxJQUFJLE9BQU8sQ0FBQyxVQUFBLE9BQU8sSUFBSSxPQUFBLE9BQU8sQ0FBQyxPQUFPLENBQUMsRUFBaEIsQ0FBZ0IsQ0FBQyxDQUFDO0lBQ2xELENBQUM7SUFFTyxpREFBYyxHQUF0QixVQUF1QixJQUFlO1FBQ3BDLE9BQU8sSUFBSSxDQUFDLFFBQVEsQ0FBQyxrQkFBa0IsQ0FBQyxJQUFJLENBQUMsQ0FBQztJQUNoRCxDQUFDO0lBckJVLHdCQUF3QjtRQURwQyxVQUFVLEVBQUU7UUFFOEIsbUJBQUEsTUFBTSxDQUFDLFVBQVUsQ0FBQyxDQUFBO2lEQUE3QixRQUFRO09BRDNCLHdCQUF3QixDQXNCcEM7SUFBRCwrQkFBQztDQUFBLEFBdEJELElBc0JDO1NBdEJZLHdCQUF3QiIsInNvdXJjZXNDb250ZW50IjpbIi8qKlxuICogQGxpY2Vuc2VcbiAqIENvcHlyaWdodCBHb29nbGUgTExDIEFsbCBSaWdodHMgUmVzZXJ2ZWQuXG4gKlxuICogVXNlIG9mIHRoaXMgc291cmNlIGNvZGUgaXMgZ292ZXJuZWQgYnkgYW4gTUlULXN0eWxlIGxpY2Vuc2UgdGhhdCBjYW4gYmVcbiAqIGZvdW5kIGluIHRoZSBMSUNFTlNFIGZpbGUgYXQgaHR0cHM6Ly9hbmd1bGFyLmlvL2xpY2Vuc2VcbiAqL1xuaW1wb3J0IHtcbiAgSW5qZWN0YWJsZSxcbiAgTmdNb2R1bGVGYWN0b3J5TG9hZGVyLFxuICBJbmplY3Rpb25Ub2tlbixcbiAgTmdNb2R1bGVGYWN0b3J5LFxuICBJbmplY3QsXG4gIFR5cGUsXG4gIENvbXBpbGVyXG59IGZyb20gJ0Bhbmd1bGFyL2NvcmUnO1xuaW1wb3J0IHtNb2R1bGVNYXB9IGZyb20gJy4vbW9kdWxlLW1hcCc7XG5cbi8qKlxuICogVG9rZW4gdXNlZCBieSB0aGUgTW9kdWxlTWFwTmdGYWN0b3J5TG9hZGVyIHRvIGxvYWQgbW9kdWxlc1xuICovXG5leHBvcnQgY29uc3QgTU9EVUxFX01BUDogSW5qZWN0aW9uVG9rZW48TW9kdWxlTWFwPiA9IG5ldyBJbmplY3Rpb25Ub2tlbignTU9EVUxFX01BUCcpO1xuXG4vKipcbiAqIE5nTW9kdWxlRmFjdG9yeUxvYWRlciB3aGljaCBkb2VzIG5vdCBsYXp5IGxvYWRcbiAqL1xuQEluamVjdGFibGUoKVxuZXhwb3J0IGNsYXNzIE1vZHVsZU1hcE5nRmFjdG9yeUxvYWRlciBpbXBsZW1lbnRzIE5nTW9kdWxlRmFjdG9yeUxvYWRlciB7XG4gIGNvbnN0cnVjdG9yKHByaXZhdGUgY29tcGlsZXI6IENvbXBpbGVyLCBASW5qZWN0KE1PRFVMRV9NQVApIHByaXZhdGUgbW9kdWxlTWFwOiBNb2R1bGVNYXApIHsgfVxuXG4gIGxvYWQobG9hZENoaWxkcmVuU3RyaW5nOiBzdHJpbmcpOiBQcm9taXNlPE5nTW9kdWxlRmFjdG9yeTxhbnk+PiB7XG4gICAgY29uc3Qgb2ZmbGluZU1vZGUgPSB0aGlzLmNvbXBpbGVyIGluc3RhbmNlb2YgQ29tcGlsZXI7XG4gICAgY29uc3QgdHlwZSA9IHRoaXMubW9kdWxlTWFwW2xvYWRDaGlsZHJlblN0cmluZ107XG5cbiAgICBpZiAoIXR5cGUpIHtcbiAgICAgIHRocm93IG5ldyBFcnJvcihgJHtsb2FkQ2hpbGRyZW5TdHJpbmd9IGRpZCBub3QgZXhpc3QgaW4gdGhlIE1PRFVMRV9NQVBgKTtcbiAgICB9XG5cbiAgICByZXR1cm4gb2ZmbGluZU1vZGUgP1xuICAgICAgdGhpcy5sb2FkRmFjdG9yeSg8TmdNb2R1bGVGYWN0b3J5PGFueT4+IHR5cGUpIDogdGhpcy5sb2FkQW5kQ29tcGlsZSg8VHlwZTxhbnk+PiB0eXBlKTtcbiAgfVxuXG4gIHByaXZhdGUgbG9hZEZhY3RvcnkoZmFjdG9yeTogTmdNb2R1bGVGYWN0b3J5PGFueT4pOiBQcm9taXNlPE5nTW9kdWxlRmFjdG9yeTxhbnk+PiB7XG4gICAgcmV0dXJuIG5ldyBQcm9taXNlKHJlc29sdmUgPT4gcmVzb2x2ZShmYWN0b3J5KSk7XG4gIH1cblxuICBwcml2YXRlIGxvYWRBbmRDb21waWxlKHR5cGU6IFR5cGU8YW55Pik6IFByb21pc2U8TmdNb2R1bGVGYWN0b3J5PGFueT4+IHtcbiAgICByZXR1cm4gdGhpcy5jb21waWxlci5jb21waWxlTW9kdWxlQXN5bmModHlwZSk7XG4gIH1cbn1cbiJdfQ==