package deobfuscator

const base64 = `var Base64={_keyStr:"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=",encode:function(r){var t,e,o,a,n,h,c,d="",i=0;for(r=Base64._utf8_encode(r);i<r.length;)a=(t=r.charCodeAt(i++))>>2,n=(3&t)<<4|(e=r.charCodeAt(i++))>>4,h=(15&e)<<2|(o=r.charCodeAt(i++))>>6,c=63&o,isNaN(e)?h=c=64:isNaN(o)&&(c=64),d=d+this._keyStr.charAt(a)+this._keyStr.charAt(n)+this._keyStr.charAt(h)+this._keyStr.charAt(c);return d},decode:function(r){var t,e,o,a,n,h,c="",d=0;for(r=r.replace(/[^A-Za-z0-9\+\/\=]/g,"");d<r.length;)t=this._keyStr.indexOf(r.charAt(d++))<<2|(a=this._keyStr.indexOf(r.charAt(d++)))>>4,e=(15&a)<<4|(n=this._keyStr.indexOf(r.charAt(d++)))>>2,o=(3&n)<<6|(h=this._keyStr.indexOf(r.charAt(d++))),c+=String.fromCharCode(t),64!=n&&(c+=String.fromCharCode(e)),64!=h&&(c+=String.fromCharCode(o));return c=Base64._utf8_decode(c)},_utf8_encode:function(r){for(var t="",e=0;e<r.length;e++){var o=r.charCodeAt(e);o<128?t+=String.fromCharCode(o):o>127&&o<2048?(t+=String.fromCharCode(o>>6|192),t+=String.fromCharCode(63&o|128)):(t+=String.fromCharCode(o>>12|224),t+=String.fromCharCode(o>>6&63|128),t+=String.fromCharCode(63&o|128))}return t},_utf8_decode:function(r){for(var t="",e=0,o=c1=c2=0;e<r.length;)(o=r.charCodeAt(e))<128?(t+=String.fromCharCode(o),e++):o>191&&o<224?(c2=r.charCodeAt(e+1),t+=String.fromCharCode((31&o)<<6|63&c2),e+=2):(c2=r.charCodeAt(e+1),c3=r.charCodeAt(e+2),t+=String.fromCharCode((15&o)<<12|(63&c2)<<6|63&c3),e+=3);return t}};function btoa(r){return Base64.encode(r.toString())}function atob(r){return Base64.decode(r.toString())}`

const window = `var window = {
	navigator: {
		userAgent: "%s",
	}
};`
