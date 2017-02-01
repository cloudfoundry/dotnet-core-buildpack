module.exports = function (callback, first, second) {
    var result = first + second;
    callback(/* error */ null, result);
};
