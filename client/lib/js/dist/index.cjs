'use strict';

var minimatch = require('minimatch');
var lodash = require('lodash');
var moment = require('moment');
var numeral = require('numeral');
var hr = require('hex-rgb');
var fs = require('fs');
var axios = require('axios');

// const iso8601check = /^([\\+-]?\d{4}(?!\d{2}\b))((-?)((0[1-9]|1[0-2])(\3([12]\d|0[1-9]|3[01]))?|W([0-4]\d|5[0-2])(-?[1-7])?|(00[1-9]|0[1-9]\d|[12]\d{2}|3([0-5]\d|6[1-6])))([T\s]((([01]\d|2[0-3])((:?)[0-5]\d)?|24:?00)([.,]\d+(?!:))?)?(\17[0-5]\d([.,]\d+)?)?([zZ]|([\\+-])([01]\d|2[0-3]):?([0-5]\d)?)?)?$/
/**
 * Reasons behind this small snippet:
 *  - backend is using uint64 as prefered type for handling CortezaID (of all things)
 *  - JavaScript can not (without external help) deal with uint64
 *  - Backend's JSON marshaller converts uint64 to string
 *  - Backend's JSON unmarshaller raises error when given anything but string with a number inside
 *
 *  Until this last thing is fixed or has a proper workaround, we're stuck with this.
 */
const NoID = '0';
function ISO8601Date(ts) {
    if (ts instanceof Date) {
        return ts;
    }
    if (typeof ts === 'string' || typeof ts === 'number') {
        return new Date(ts);
    }
    return undefined;
}
/**
 * Tests if a given value looks like corteza ID
 * @param ID
 * @constructor
 */
function IsCortezaID(ID) {
    if (typeof ID !== 'string') {
        return false;
    }
    if (!/^\d+$/.test(ID)) {
        return false;
    }
    return true;
}
/**
 * @return {string}
 */
function CortezaID(value) {
    if (!value) {
        return NoID;
    }
    if (typeof value === 'number') {
        return value.toString();
    }
    if (IsCortezaID(value)) {
        return value;
    }
    throw new Error('Invalid CortezaID value');
}
/**
 * Apply takes all given props, their values (from src) and assignes them to props (on dst)
 *
 * A casting function can be used (see ApplyCaster) to modify the values before assigning them
 */
function Apply(dst, src, cast, ...props) {
    if (typeof cast !== 'function') {
        // Handle case where we do not use caster
        props.unshift(cast);
        // and use String as a caster, effectively forcing
        // cast-to-string on all applied values to
        cast = String;
    }
    if (typeof src !== 'object') {
        return;
    }
    props.forEach(prop => {
        // prop must exist on dst
        if (!Object.prototype.hasOwnProperty.call(dst, prop)) {
            return;
        }
        // prop must exist on src
        if (!Object.prototype.hasOwnProperty.call(src, prop)) {
            return;
        }
        // sProp is prop from source
        const sProp = prop;
        // value on src should be defined
        if (!src || src[sProp] === undefined || src[sProp] === null) {
            return;
        }
        // Cast value from src to type of value from prop on dst
        let val = src[sProp];
        // Run value through cast fn
        val = cast(val);
        if (val === undefined) {
            return;
        }
        // Assign (valid ony) value to dst
        dst[prop] = val;
    });
}
function ApplyWhitelisted(dst, src, whitelist, ...props) {
    if (typeof src !== 'object') {
        return;
    }
    props.forEach(prop => {
        // prop must exist on dst
        if (!Object.prototype.hasOwnProperty.call(dst, prop)) {
            return;
        }
        // prop must exist on src
        if (!Object.prototype.hasOwnProperty.call(src, prop)) {
            return;
        }
        // sProp is prop from source
        const sProp = prop;
        // value on src should be defined
        if (!src || src[sProp] === undefined) {
            return;
        }
        // Cast value from src to type of value from prop on dst
        const val = src[sProp];
        if (whitelist.includes(val)) {
            dst[prop] = val;
        }
    });
}

const IsOf = (v, ...props) => {
    if (!v || typeof v !== 'object') {
        return false;
    }
    for (const prop of props) {
        if (!Object.prototype.hasOwnProperty.call(v, prop)) {
            return false;
        }
    }
    return true;
};
// eslint-disable-next-line valid-typeof
const every = (a, t) => Array.isArray(a) && a.every(i => typeof i === t);
const AreStrings = (a) => every(a, 'string');
const AreObjects = (a) => every(a, 'object');
function AreObjectsOf(a, ...props) {
    if (!a || !Array.isArray(a)) {
        return false;
    }
    if (a.length === 0) {
        return true;
    }
    return AreObjects(a) && a.every(i => IsOf(i, ...props));
}

class Equal {
    constructor(name, vv, not = false) {
        this.name = name;
        this.values = vv;
        this.not = not;
    }
    Name() {
        return this.name;
    }
    Values() {
        return this.values;
    }
    Match(value) {
        for (const v of this.values) {
            if (value === v) {
                return !this.not;
            }
        }
        return this.not;
    }
}
/**
 * Handle glob-like pattern matching
 *
 * See: https://github.com/isaacs/minimatch
 */
class Like extends Equal {
    constructor(name, vv, not = false) {
        super(name, vv.map(v => v.replace('%', '*').replace('_', '?')), not);
    }
    Match(value) {
        for (const v of this.values) {
            if (minimatch.minimatch(value, v)) {
                return !this.not;
            }
        }
        return this.not;
    }
}
/**
 * Regex matcher
 */
class Match extends Equal {
    constructor(name, vv, not = false) {
        super(name, vv, not);
        this.re = vv.map(v => new RegExp(v));
    }
    Match(value) {
        for (const re of this.re) {
            if (re.test(value)) {
                return !this.not;
            }
        }
        return this.not;
    }
}
function ConstraintMaker(c) {
    if (!IsOf(c, 'value')) {
        throw new Error('invalid constraint input');
    }
    const { name = '', op = '', value } = c;
    switch (op.toLowerCase()) {
        case '':
        case 'eq':
        case '=':
        case '==':
        case '===':
            return new Equal(name, value);
        case 'not eq':
        case 'ne':
        case '!=':
        case '!==':
            return new Equal(name, value, true);
        case 'like':
            return new Like(name, value);
        case 'not like':
            return new Like(name, value, true);
        case '~':
            return new Match(name, value);
        case '!~':
            return new Match(name, value, true);
        default:
            throw new Error('unsupported constraint operator');
    }
}

/******************************************************************************
Copyright (c) Microsoft Corporation.

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
PERFORMANCE OF THIS SOFTWARE.
***************************************************************************** */
/* global Reflect, Promise, SuppressedError, Symbol, Iterator */


function __rest(s, e) {
    var t = {};
    for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p) && e.indexOf(p) < 0)
        t[p] = s[p];
    if (s != null && typeof Object.getOwnPropertySymbols === "function")
        for (var i = 0, p = Object.getOwnPropertySymbols(s); i < p.length; i++) {
            if (e.indexOf(p[i]) < 0 && Object.prototype.propertyIsEnumerable.call(s, p[i]))
                t[p[i]] = s[p[i]];
        }
    return t;
}

function __awaiter(thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
}

typeof SuppressedError === "function" ? SuppressedError : function (error, suppressed, message) {
    var e = new Error(message);
    return e.name = "SuppressedError", e.error = error, e.suppressed = suppressed, e;
};

class Handler {
    constructor(h, t) {
        this.handle = h;
        this.eventTypes = t.eventTypes;
        this.resourceTypes = t.resourceTypes;
        this.weight = t.weight || 0;
        // @todo parse constraints to constraint matchers
        this.constraints = t.constraints ? t.constraints.map(ConstraintMaker) : [];
        this.scriptName = t.scriptName;
    }
    /**
     * Match this handler with a given event - type, resource, constraints + scriptName when ManualEvent
     *
     * @param {Event} ev
     * @return bool
     */
    Match(ev, script) {
        if (!this.eventTypes.includes(ev.eventType)) {
            return false;
        }
        if (!this.resourceTypes.includes(ev.resourceType)) {
            return false;
        }
        if (script && this.scriptName !== script) {
            return false;
        }
        if (ev.match) {
            // Event should match all trigger's constraints
            for (const c of this.constraints) {
                if (!ev.match(c)) {
                    return false;
                }
            }
        }
        return true;
    }
    Handle(ev) {
        return this.handle(ev);
    }
}

const onManual = 'onManual';
function scriptSorter(a, b) {
    return a.weight - b.weight;
}
function GenericEventMaker(t, eventType, match, args) {
    return {
        resourceType: t.resourceType,
        eventType,
        match,
        args,
    };
}

/**
 * EventBus handles all Corteza events on browser (!! not on corredor server !!)
 *
 * Flow #1
 *  1. Corredor prepares a bundle that is loaded on a client
 *  2. Bundle provides a "callback" function that accepts EventBus object +
 *     all context information and configuration that is needed for
 *     handler registration
 * 3a. When a "Corteza event" is dispatched (Dispatch())
 *     event-bus searches for handler and passes on event information
 * 3b. When manual event is executed (Exec())
 *     event-bus searches for handler and passes on event information
 *
 * Flow #2
 *   1. When web application is initialized, it should register all
 *      explicit server scripts
 *   2. These server scripts are wrapped with a handlerFn that forwards call
 *      to the API (there, request is passed to the Corredor where it's executed)
 *
 */
/**
 * EventBus for event dispatching and handling
 *
 * Since we have much shorter execution path here than we have in case of server scripts,
 * we can afford some optimisation (in comparison to backend's pkg/eventbus)
 */
class EventBus {
    constructor(opt) {
        this.handlers = [];
        this.pairs = (opt === null || opt === void 0 ? void 0 : opt.pairs) || {};
        this.strict = !!(opt === null || opt === void 0 ? void 0 : opt.strict);
        this.verbose = !!(opt === null || opt === void 0 ? void 0 : opt.verbose);
    }
    /**
     * Dispatches event and sequentially calls all handlers.
     *
     * Handling handler results works a bit different then on backend.
     * Scripts executed with handlers have DIRECT access to values passed (by reference)
     * as arguments via event so there's no need to do an explicit return
     *
     * @param {Event} ev Event to dispatch
     */
    Dispatch(ev, script) {
        return __awaiter(this, void 0, void 0, function* () {
            if (this.verbose)
                console.debug('EventBus: event dispatched', { ev, script });
            if (script) {
                if (ev.eventType !== onManual) {
                    console.warn('EventBus: explicit events require onManual event type', ev);
                    return null;
                }
            }
            else {
                if (ev.eventType === onManual) {
                    console.warn('EventBus: implicit events can not define onManual event type', ev);
                    return null;
                }
            }
            this.checkPairs([ev.resourceType], [ev.eventType]);
            const matched = this.find(ev, script);
            if (matched.length === 0) {
                if (this.verbose)
                    console.debug('EventBus: no handlers found', { ev, script, registeredHandlers: this.handlers.length });
                return null;
            }
            if (script) {
                // When executing a specific script,
                // make sure we do not run it multiple times.
                matched.splice(1);
            }
            try {
                for (const t of matched) {
                    if (this.verbose)
                        console.debug('EventBus: handling event', { ev, trigger: t, script });
                    const result = yield t.Handle(ev);
                    if (result === false) {
                        return Promise.reject(new Error('aborted'));
                    }
                }
            }
            catch (err) {
                return Promise.reject(err);
            }
            return null;
        });
    }
    /**
     * Filters and sorts all handlers by event & constraints
     */
    find(ev, script) {
        return this.handlers
            .filter(t => t.Match(ev, script))
            .sort(scriptSorter);
    }
    /**
     * Registers Event handler
     *
     * @param handler Handler function
     * @param trigger Trigger definition
     */
    Register(handler, trigger) {
        if (this.verbose)
            console.debug('EventBus: event handler registration for', trigger.scriptName, { trigger });
        this.handlers.push(new Handler(handler, trigger));
        return this;
    }
    /**
     * Unregisters all handlers
     */
    UnregisterAll() {
        this.handlers = [];
        return this;
    }
    checkPairs(resourceTypes, eventTypes) {
        if (this.pairs === undefined || !this.strict) {
            return;
        }
        resourceTypes.forEach(resourceType => {
            const wket = this.pairs[resourceType];
            if (wket === undefined) {
                throw new TypeError('unknown resource type "' + resourceType + "'");
            }
            eventTypes.forEach(eventTypes => {
                if (!wket.includes(eventTypes)) {
                    throw new TypeError('unknown event type "' + eventTypes + '" for "' + resourceType + '" resource type');
                }
            });
        });
    }
}

var index$8 = /*#__PURE__*/Object.freeze({
  __proto__: null,
  ConstraintMaker: ConstraintMaker,
  EventBus: EventBus
});

/**
 * Script executor
 *
 * @param script - Script to be executed
 * @param args - Arguments for the script
 * @param ctx - Exec context (exec function's 2nd param)
 */
function Exec(script, args, ctx) {
    return __awaiter(this, void 0, void 0, function* () {
        try {
            // Wrap exec() with Promise.resolve - we do not know if function is async or not.
            return Promise.resolve(script.exec(args, ctx)).then((rval) => {
                let result = {};
                if (rval === false) {
                    // Abort when returning false!
                    throw new Error('Aborted');
                }
                if (typeof rval === 'object' && rval && rval.constructor.name === 'Object') {
                    // Expand returned values into result if function returned a plain javascript object
                    result = Object.assign({}, rval);
                }
                else if (rval !== undefined) {
                    // If anything usable was returned, stack it under 'result' property
                    result = { result: rval };
                }
                // Wrap returning value
                return result;
            });
        }
        catch (e) {
            return Promise.reject(e);
        }
    });
}

function kv(a) { return a; }
/**
 * Extracts ID-like (numeric) value from string or object
 *
 * @param value - that stores ID in some way
 * @param prop - possible key lookup
 */
function extractID(value, prop) {
    if (value && typeof value === 'object') {
        if (!prop || !Object.prototype.hasOwnProperty.call(value, prop)) {
            return NoID;
        }
        value = value[prop];
    }
    return CortezaID(value);
}
function isFresh(ID) {
    return !ID || ID === NoID;
}
function genericPermissionUpdater(API, rules) {
    const g = rules.reduce((acc, p) => {
        if (!acc[p.role.roleID]) {
            acc[p.role.roleID] = [];
        }
        acc[p.role.roleID].push({
            resource: p.resource.resourceID,
            operation: p.operation,
            access: p.access,
        });
        return acc;
    }, {});
    // @todo should return promise and stack all these into Promise.all()
    Object.keys(g).forEach((roleID) => __awaiter(this, void 0, void 0, function* () {
        // permissions grouped per role
        yield API.permissionsUpdate({ roleID, rules: g[roleID] });
    }));
}

class User {
    constructor(u) {
        this.userID = NoID;
        this.handle = '';
        this.username = '';
        this.userGroupID = NoID;
        this.email = '';
        this.name = '';
        this.emailConfirmed = false;
        this.labels = {};
        this.meta = {
            preferredLanguage: 'en',
            securityPolicy: {
                mfa: {
                    enforcedEmailOTP: false,
                    enforcedTOTP: false,
                },
            },
            avatarID: NoID,
            avatarKind: '',
            avatarColor: '',
            avatarBgColor: '',
            theme: '',
        };
        this.canGrant = false;
        this.canUpdateUser = false;
        this.canDeleteUser = false;
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.suspendedAt = undefined;
        this.apply(u);
    }
    apply(u) {
        Apply(this, u, CortezaID, 'userID', 'userGroupID');
        Apply(this, u, String, 'handle', 'username', 'email', 'name');
        Apply(this, u, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt', 'suspendedAt');
        Apply(this, u, Boolean, 'emailConfirmed', 'canGrant', 'canUpdateUser', 'canDeleteUser');
        if (u === null || u === void 0 ? void 0 : u.roles) {
            this.roles = [];
            if (AreStrings(u.roles)) {
                this.roles = u.roles;
            }
        }
        if (IsOf(u, 'meta')) {
            this.meta = Object.assign({}, u.meta);
        }
        if (IsOf(u, 'labels')) {
            this.labels = Object.assign({}, u.labels);
        }
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.userID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'system:user';
    }
    get fts() {
        return [
            this.name,
            this.username,
            this.handle,
            this.email,
            this.userID,
        ].join(' ').toLocaleLowerCase();
    }
    clone() {
        return new User(JSON.parse(JSON.stringify(this)));
    }
    properties() {
        return [
            'userID',
            'handle',
            'username',
            'email',
            'userGroupID',
            'name',
            'emailConfirmed',
            'labels',
            'meta',
            'canGrant',
            'canUpdateUser',
            'canDeleteUser',
            'createdAt',
            'updatedAt',
            'deletedAt',
            'suspendedAt',
            'roles',
        ];
    }
}

const defaultMeta$1 = {
    description: '',
    context: {
        resourceTypes: [],
        expr: '',
    },
};
class Role {
    constructor(r) {
        this.roleID = NoID;
        this.name = '';
        this.handle = '';
        this.members = [];
        this.labels = {};
        this.meta = Object.assign({}, defaultMeta$1);
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.archivedAt = undefined;
        this.isSystem = false;
        this.isClosed = false;
        this.isBypass = false;
        this.canGrant = false;
        this.canUpdateRole = false;
        this.canDeleteRole = false;
        this.canManageMembersOnRole = false;
        this.apply(r);
    }
    apply(r) {
        Apply(this, r, CortezaID, 'roleID');
        Apply(this, r, String, 'name', 'handle');
        if (r === null || r === void 0 ? void 0 : r.members) {
            this.members = [];
            if (AreStrings(r.members)) {
                this.members = r.members;
            }
        }
        if (IsOf(r, 'meta')) {
            this.meta = Object.assign({}, r.meta);
        }
        if (!this.meta) {
            this.meta = Object.assign({}, defaultMeta$1);
        }
        if (!this.meta.context) {
            this.meta.context = Object.assign({}, defaultMeta$1.context);
        }
        if (IsOf(r, 'labels')) {
            this.labels = Object.assign({}, r.labels);
        }
        Apply(this, r, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt', 'archivedAt');
        Apply(this, r, Boolean, 'isSystem', 'isClosed', 'isBypass', 'canGrant', 'canUpdateRole', 'canDeleteRole', 'canManageMembersOnRole');
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.roleID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'system:role';
    }
    get isContext() {
        var _a, _b, _c, _d, _e, _f;
        return ((_c = (_b = (_a = this.meta) === null || _a === void 0 ? void 0 : _a.context) === null || _b === void 0 ? void 0 : _b.expr) === null || _c === void 0 ? void 0 : _c.length) > 0 || ((_f = (_e = (_d = this.meta) === null || _d === void 0 ? void 0 : _d.context) === null || _e === void 0 ? void 0 : _e.resourceTypes) === null || _f === void 0 ? void 0 : _f.length) > 0;
    }
    clone() {
        return new Role(JSON.parse(JSON.stringify(this)));
    }
}

class Application {
    constructor(r) {
        this.applicationID = undefined;
        this.name = '';
        this.ownerID = 0;
        this.enabled = false;
        this.weight = 0;
        this.unify = {
            name: '',
            listed: false,
            url: '',
            config: '',
            iconID: NoID,
            logoID: NoID,
        };
        this.canGrant = true;
        this.canUpdateApplication = true;
        this.canDeleteApplication = true;
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.apply(r);
    }
    apply(r) {
        Apply(this, r, CortezaID, 'applicationID');
        Apply(this, r, String, 'name');
        Apply(this, r, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, r, Number, 'weight', 'ownerID');
        Apply(this, r, Boolean, 'enabled', 'canGrant', 'canUpdateApplication', 'canDeleteApplication');
        if (r && IsOf(r, 'unify')) {
            this.unify = r.unify;
        }
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.applicationID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'system:application';
    }
    clone() {
        return new Application(JSON.parse(JSON.stringify(this)));
    }
}

const { merge: merge$d } = lodash;
class SinkRequest {
    constructor(r = {}) {
        this.method = '';
        this.path = '';
        this.host = '';
        this.header = {};
        this.query = {};
        this.postForm = {};
        this.username = '';
        this.password = '';
        this.remoteAddress = '';
        this.rawBody = '';
        merge$d(this, r);
    }
}
class SinkResponse {
    constructor(r = {}) {
        this.status = 200;
        this.header = {};
        merge$d(this, r);
    }
}

const defaultMeta = {
    description: '',
    short: '',
};
const defaultConfig = {
    path: [],
};
class UserGroup {
    constructor(u) {
        this.userGroupID = NoID;
        this.handle = '';
        this.isRoot = false;
        this.config = Object.assign({}, defaultConfig);
        this.meta = Object.assign({}, defaultMeta);
        this.labels = {};
        this.canGrant = false;
        this.canUpdateUserGroup = false;
        this.canDeleteUserGroup = false;
        this.canManageMembersOnUserGroup = false;
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.suspendedAt = undefined;
        this.apply(u);
    }
    apply(u) {
        Apply(this, u, CortezaID, 'userGroupID');
        Apply(this, u, String, 'handle');
        Apply(this, u, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt', 'suspendedAt');
        Apply(this, u, Boolean, 'isRoot', 'canGrant', 'canUpdateUserGroup', 'canDeleteUserGroup', 'canManageMembersOnUserGroup');
        if (u === null || u === void 0 ? void 0 : u.roles) {
            this.roles = [];
            if (AreStrings(u.roles)) {
                this.roles = u.roles;
            }
        }
        if (IsOf(u, 'config')) {
            this.config = Object.assign({}, u.config);
        }
        if (!this.config) {
            this.config = Object.assign({}, defaultConfig);
        }
        if (IsOf(u, 'meta')) {
            this.meta = Object.assign({}, u.meta);
        }
        if (!this.meta) {
            this.meta = Object.assign({}, defaultMeta);
        }
        if (IsOf(u, 'labels')) {
            this.labels = Object.assign({}, u.labels);
        }
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.userGroupID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'system:user-group';
    }
    get fts() {
        return [
            this.meta.short,
            this.handle,
            this.userGroupID,
        ].join(' ').toLocaleLowerCase();
    }
    clone() {
        return new UserGroup(JSON.parse(JSON.stringify(this)));
    }
    properties() {
        return [
            'userGroupID',
            'handle',
            'labels',
            'isRoot',
            'canGrant',
            'canUpdateUserGroup',
            'canDeleteUserGroup',
            'canManageMembersOnUserGroup',
            'createdAt',
            'updatedAt',
            'deletedAt',
            'suspendedAt',
            'roles',
        ];
    }
}

const { merge: merge$c } = lodash;
class Reminder {
    constructor(r) {
        this.reminderID = NoID;
        this.resource = NoID;
        this.payload = {};
        this.snoozeCount = 0;
        this.assignedTo = NoID;
        this.assignedBy = NoID;
        this.assignedAt = undefined;
        this.dismissedBy = NoID;
        this.dismissedAt = undefined;
        this.remindAt = undefined;
        this.createdAt = undefined;
        this.processed = false;
        this.actions = {};
        this.options = {};
        this.apply(r);
    }
    apply(r) {
        if (!r)
            return;
        Apply(this, r, CortezaID, 'reminderID');
        Apply(this, r, Number, 'snoozeCount');
        Apply(this, r, CortezaID, 'assignedTo', 'assignedBy', 'dismissedBy');
        // @todo actions, options, payload... all 3?
        this.payload = merge$c({}, this.payload, r.payload);
        this.actions = merge$c({}, this.actions, r.actions);
        this.options = merge$c({}, this.options, r.options);
        Apply(this, r, ISO8601Date, 'assignedAt', 'dismissedAt', 'remindAt', 'createdAt');
        Apply(this, r, Boolean, 'processed');
    }
}

class Template {
    constructor(r) {
        this.templateID = NoID;
        this.handle = '';
        this.language = '';
        this.type = 'text/html';
        this.partial = false;
        this.meta = {
            short: '',
            description: '',
        };
        this.template = '';
        this.labels = {};
        this.ownerID = NoID;
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.lastUsedAt = undefined;
        this.canDeleteTemplate = false;
        this.apply(r);
    }
    apply(r) {
        Apply(this, r, CortezaID, 'templateID', 'ownerID');
        Apply(this, r, String, 'handle', 'language', 'type', 'template');
        Apply(this, r, Boolean, 'partial', 'canDeleteTemplate');
        if (r && IsOf(r, 'meta')) {
            this.meta = r.meta;
        }
        Apply(this, r, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt', 'lastUsedAt');
        if (IsOf(r, 'labels')) {
            this.labels = Object.assign({}, r.labels);
        }
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.templateID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'system:template';
    }
    clone() {
        return new Template(JSON.parse(JSON.stringify(this)));
    }
}

const { merge: merge$b } = lodash;
class DisplayElement {
    constructor(de = {}) {
        this.elementID = NoID;
        this.name = '';
        this.description = '';
        this.options = {};
        this.meta = {
            size: undefined,
        };
        this.kind = '';
        this.apply(de);
    }
    apply(de) {
        if (!de)
            return;
        Apply(this, de, String, 'name', 'description');
        Apply(this, de, CortezaID, 'elementID');
        if (de.options) {
            this.options = merge$b({}, this.options, de.options);
        }
        if (de.meta) {
            this.meta = merge$b({}, this.meta, de.meta);
        }
    }
}
const Registry$2 = new Map();

const kind$w = 'Chart';
class ChartOptions {
    constructor(o = {}) {
        this.title = '';
        this.type = 'bar';
        this.colorScheme = '';
        this.noAnimation = false;
        this.source = '';
        this.datasources = [];
        this.xAxis = {
            type: '',
            skipMissing: true,
            labelRotation: 0,
        };
        this.yAxis = {
            type: 'linear',
            position: 'left',
            labelPosition: 'end',
            labelRotation: 0,
            beginAtZero: true,
        };
        this.legend = {
            hide: false,
            orientation: 'horizontal',
            align: 'center',
            scrollable: true,
            position: {
                default: true,
                top: undefined,
                right: undefined,
                bottom: undefined,
                left: undefined,
            },
        };
        this.tooltips = {
            showAlways: false,
        };
        this.offset = {
            default: true,
            top: undefined,
            right: undefined,
            bottom: undefined,
            left: undefined,
        };
        if (!o)
            return;
        Apply(this, o, String, 'title', 'type', 'colorScheme', 'source');
        Apply(this, o, Boolean, 'noAnimation');
        if (o.datasources) {
            this.datasources = o.datasources;
        }
        if (o.xAxis) {
            this.xAxis = Object.assign(Object.assign({}, this.xAxis), o.xAxis);
        }
        if (o.yAxis) {
            this.yAxis = Object.assign(Object.assign({}, this.yAxis), o.yAxis);
        }
        if (o.legend) {
            this.legend = Object.assign(Object.assign({}, this.legend), o.legend);
        }
        if (o.tooltips) {
            this.tooltips = Object.assign(Object.assign({}, this.tooltips), o.tooltips);
        }
        if (o.offset) {
            this.offset = Object.assign(Object.assign({}, this.offset), o.offset);
        }
    }
}
const ChartOptionsRegistry = new Map();
function ChartOptionsMaker(options) {
    const { type } = options;
    if (type) {
        const ChartOptionsTemp = ChartOptionsRegistry.get(type);
        if (ChartOptionsTemp === undefined) {
            throw new Error(`unknown chart type '${type}'`);
        }
        if (options instanceof ChartOptions) {
            // Get rid of the references
            options = JSON.parse(JSON.stringify(options));
        }
        return new ChartOptionsTemp(options);
    }
    else {
        throw new Error('no chart type');
    }
}
class DisplayElementChart extends DisplayElement {
    constructor(i) {
        super(i);
        this.kind = kind$w;
        this.options = ChartOptionsMaker({ type: 'bar' });
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        this.options = ChartOptionsMaker(o);
    }
    reportDefinitions(definition = {}) {
        if (typeof this.options.source === 'object') {
            // @todo allow implicit sources
            throw new Error('chart source must be provided as a reference');
        }
        const dataframes = [];
        this.options.datasources.forEach(({ name = '', filter, sort }) => {
            var _a;
            const df = {
                name: this.elementID,
                source: this.options.source,
                ref: name,
                filter,
                sort,
            };
            const relatedDefinition = definition[name];
            if (relatedDefinition) {
                df.sort = (relatedDefinition.sort ? relatedDefinition.sort : sort) || undefined;
                if (relatedDefinition.filter && ((_a = relatedDefinition.filter) === null || _a === void 0 ? void 0 : _a.ref)) {
                    // If element and scenario have filter AND them together
                    if (filter && filter.ref) {
                        df.filter = {
                            ref: 'and',
                            args: [
                                filter,
                                relatedDefinition.filter,
                            ],
                        };
                    }
                    else {
                        df.filter = relatedDefinition.filter;
                    }
                }
            }
            dataframes.push(df);
        });
        return { dataframes };
    }
}
Registry$2.set(kind$w, DisplayElementChart);

class Attachment {
    constructor(i, baseURL) {
        this.attachmentID = NoID;
        this.ownerID = NoID;
        this.name = '';
        this.url = '';
        this.previewUrl = '';
        this.download = '';
        this.meta = {};
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.apply(i);
        this.setBaseURL(baseURL || '');
    }
    apply(i) {
        Apply(this, i, CortezaID, 'attachmentID', 'ownerID');
        Apply(this, i, String, 'name', 'url', 'previewUrl');
        if (IsOf(i, 'meta')) {
            this.meta = Object.assign({}, i.meta);
        }
        Apply(this, i, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
    }
    setBaseURL(baseURL) {
        this.url = baseURL + this.url;
        this.previewUrl = baseURL + this.previewUrl;
        this.download = this.url + '&download=1';
    }
}

var brewer = {
    // Sequential
    YlGn3: ['#f7fcb9', '#addd8e', '#31a354'],
    YlGn4: ['#ffffcc', '#c2e699', '#78c679', '#238443'],
    YlGn5: ['#ffffcc', '#c2e699', '#78c679', '#31a354', '#006837'],
    YlGn6: ['#ffffcc', '#d9f0a3', '#addd8e', '#78c679', '#31a354', '#006837'],
    YlGn7: ['#ffffcc', '#d9f0a3', '#addd8e', '#78c679', '#41ab5d', '#238443', '#005a32'],
    YlGn8: ['#ffffe5', '#f7fcb9', '#d9f0a3', '#addd8e', '#78c679', '#41ab5d', '#238443', '#005a32'],
    YlGn9: ['#ffffe5', '#f7fcb9', '#d9f0a3', '#addd8e', '#78c679', '#41ab5d', '#238443', '#006837', '#004529'],
    YlGnBu3: ['#edf8b1', '#7fcdbb', '#2c7fb8'],
    YlGnBu4: ['#ffffcc', '#a1dab4', '#41b6c4', '#225ea8'],
    YlGnBu5: ['#ffffcc', '#a1dab4', '#41b6c4', '#2c7fb8', '#253494'],
    YlGnBu6: ['#ffffcc', '#c7e9b4', '#7fcdbb', '#41b6c4', '#2c7fb8', '#253494'],
    YlGnBu7: ['#ffffcc', '#c7e9b4', '#7fcdbb', '#41b6c4', '#1d91c0', '#225ea8', '#0c2c84'],
    YlGnBu8: ['#ffffd9', '#edf8b1', '#c7e9b4', '#7fcdbb', '#41b6c4', '#1d91c0', '#225ea8', '#0c2c84'],
    YlGnBu9: ['#ffffd9', '#edf8b1', '#c7e9b4', '#7fcdbb', '#41b6c4', '#1d91c0', '#225ea8', '#253494', '#081d58'],
    GnBu3: ['#e0f3db', '#a8ddb5', '#43a2ca'],
    GnBu4: ['#f0f9e8', '#bae4bc', '#7bccc4', '#2b8cbe'],
    GnBu5: ['#f0f9e8', '#bae4bc', '#7bccc4', '#43a2ca', '#0868ac'],
    GnBu6: ['#f0f9e8', '#ccebc5', '#a8ddb5', '#7bccc4', '#43a2ca', '#0868ac'],
    GnBu7: ['#f0f9e8', '#ccebc5', '#a8ddb5', '#7bccc4', '#4eb3d3', '#2b8cbe', '#08589e'],
    GnBu8: ['#f7fcf0', '#e0f3db', '#ccebc5', '#a8ddb5', '#7bccc4', '#4eb3d3', '#2b8cbe', '#08589e'],
    GnBu9: ['#f7fcf0', '#e0f3db', '#ccebc5', '#a8ddb5', '#7bccc4', '#4eb3d3', '#2b8cbe', '#0868ac', '#084081'],
    BuGn3: ['#e5f5f9', '#99d8c9', '#2ca25f'],
    BuGn4: ['#edf8fb', '#b2e2e2', '#66c2a4', '#238b45'],
    BuGn5: ['#edf8fb', '#b2e2e2', '#66c2a4', '#2ca25f', '#006d2c'],
    BuGn6: ['#edf8fb', '#ccece6', '#99d8c9', '#66c2a4', '#2ca25f', '#006d2c'],
    BuGn7: ['#edf8fb', '#ccece6', '#99d8c9', '#66c2a4', '#41ae76', '#238b45', '#005824'],
    BuGn8: ['#f7fcfd', '#e5f5f9', '#ccece6', '#99d8c9', '#66c2a4', '#41ae76', '#238b45', '#005824'],
    BuGn9: ['#f7fcfd', '#e5f5f9', '#ccece6', '#99d8c9', '#66c2a4', '#41ae76', '#238b45', '#006d2c', '#00441b'],
    PuBuGn3: ['#ece2f0', '#a6bddb', '#1c9099'],
    PuBuGn4: ['#f6eff7', '#bdc9e1', '#67a9cf', '#02818a'],
    PuBuGn5: ['#f6eff7', '#bdc9e1', '#67a9cf', '#1c9099', '#016c59'],
    PuBuGn6: ['#f6eff7', '#d0d1e6', '#a6bddb', '#67a9cf', '#1c9099', '#016c59'],
    PuBuGn7: ['#f6eff7', '#d0d1e6', '#a6bddb', '#67a9cf', '#3690c0', '#02818a', '#016450'],
    PuBuGn8: ['#fff7fb', '#ece2f0', '#d0d1e6', '#a6bddb', '#67a9cf', '#3690c0', '#02818a', '#016450'],
    PuBuGn9: ['#fff7fb', '#ece2f0', '#d0d1e6', '#a6bddb', '#67a9cf', '#3690c0', '#02818a', '#016c59', '#014636'],
    PuBu3: ['#ece7f2', '#a6bddb', '#2b8cbe'],
    PuBu4: ['#f1eef6', '#bdc9e1', '#74a9cf', '#0570b0'],
    PuBu5: ['#f1eef6', '#bdc9e1', '#74a9cf', '#2b8cbe', '#045a8d'],
    PuBu6: ['#f1eef6', '#d0d1e6', '#a6bddb', '#74a9cf', '#2b8cbe', '#045a8d'],
    PuBu7: ['#f1eef6', '#d0d1e6', '#a6bddb', '#74a9cf', '#3690c0', '#0570b0', '#034e7b'],
    PuBu8: ['#fff7fb', '#ece7f2', '#d0d1e6', '#a6bddb', '#74a9cf', '#3690c0', '#0570b0', '#034e7b'],
    PuBu9: ['#fff7fb', '#ece7f2', '#d0d1e6', '#a6bddb', '#74a9cf', '#3690c0', '#0570b0', '#045a8d', '#023858'],
    BuPu3: ['#e0ecf4', '#9ebcda', '#8856a7'],
    BuPu4: ['#edf8fb', '#b3cde3', '#8c96c6', '#88419d'],
    BuPu5: ['#edf8fb', '#b3cde3', '#8c96c6', '#8856a7', '#810f7c'],
    BuPu6: ['#edf8fb', '#bfd3e6', '#9ebcda', '#8c96c6', '#8856a7', '#810f7c'],
    BuPu7: ['#edf8fb', '#bfd3e6', '#9ebcda', '#8c96c6', '#8c6bb1', '#88419d', '#6e016b'],
    BuPu8: ['#f7fcfd', '#e0ecf4', '#bfd3e6', '#9ebcda', '#8c96c6', '#8c6bb1', '#88419d', '#6e016b'],
    BuPu9: ['#f7fcfd', '#e0ecf4', '#bfd3e6', '#9ebcda', '#8c96c6', '#8c6bb1', '#88419d', '#810f7c', '#4d004b'],
    RdPu3: ['#fde0dd', '#fa9fb5', '#c51b8a'],
    RdPu4: ['#feebe2', '#fbb4b9', '#f768a1', '#ae017e'],
    RdPu5: ['#feebe2', '#fbb4b9', '#f768a1', '#c51b8a', '#7a0177'],
    RdPu6: ['#feebe2', '#fcc5c0', '#fa9fb5', '#f768a1', '#c51b8a', '#7a0177'],
    RdPu7: ['#feebe2', '#fcc5c0', '#fa9fb5', '#f768a1', '#dd3497', '#ae017e', '#7a0177'],
    RdPu8: ['#fff7f3', '#fde0dd', '#fcc5c0', '#fa9fb5', '#f768a1', '#dd3497', '#ae017e', '#7a0177'],
    RdPu9: ['#fff7f3', '#fde0dd', '#fcc5c0', '#fa9fb5', '#f768a1', '#dd3497', '#ae017e', '#7a0177', '#49006a'],
    PuRd3: ['#e7e1ef', '#c994c7', '#dd1c77'],
    PuRd4: ['#f1eef6', '#d7b5d8', '#df65b0', '#ce1256'],
    PuRd5: ['#f1eef6', '#d7b5d8', '#df65b0', '#dd1c77', '#980043'],
    PuRd6: ['#f1eef6', '#d4b9da', '#c994c7', '#df65b0', '#dd1c77', '#980043'],
    PuRd7: ['#f1eef6', '#d4b9da', '#c994c7', '#df65b0', '#e7298a', '#ce1256', '#91003f'],
    PuRd8: ['#f7f4f9', '#e7e1ef', '#d4b9da', '#c994c7', '#df65b0', '#e7298a', '#ce1256', '#91003f'],
    PuRd9: ['#f7f4f9', '#e7e1ef', '#d4b9da', '#c994c7', '#df65b0', '#e7298a', '#ce1256', '#980043', '#67001f'],
    OrRd3: ['#fee8c8', '#fdbb84', '#e34a33'],
    OrRd4: ['#fef0d9', '#fdcc8a', '#fc8d59', '#d7301f'],
    OrRd5: ['#fef0d9', '#fdcc8a', '#fc8d59', '#e34a33', '#b30000'],
    OrRd6: ['#fef0d9', '#fdd49e', '#fdbb84', '#fc8d59', '#e34a33', '#b30000'],
    OrRd7: ['#fef0d9', '#fdd49e', '#fdbb84', '#fc8d59', '#ef6548', '#d7301f', '#990000'],
    OrRd8: ['#fff7ec', '#fee8c8', '#fdd49e', '#fdbb84', '#fc8d59', '#ef6548', '#d7301f', '#990000'],
    OrRd9: ['#fff7ec', '#fee8c8', '#fdd49e', '#fdbb84', '#fc8d59', '#ef6548', '#d7301f', '#b30000', '#7f0000'],
    YlOrRd3: ['#ffeda0', '#feb24c', '#f03b20'],
    YlOrRd4: ['#ffffb2', '#fecc5c', '#fd8d3c', '#e31a1c'],
    YlOrRd5: ['#ffffb2', '#fecc5c', '#fd8d3c', '#f03b20', '#bd0026'],
    YlOrRd6: ['#ffffb2', '#fed976', '#feb24c', '#fd8d3c', '#f03b20', '#bd0026'],
    YlOrRd7: ['#ffffb2', '#fed976', '#feb24c', '#fd8d3c', '#fc4e2a', '#e31a1c', '#b10026'],
    YlOrRd8: ['#ffffcc', '#ffeda0', '#fed976', '#feb24c', '#fd8d3c', '#fc4e2a', '#e31a1c', '#b10026'],
    YlOrRd9: ['#ffffcc', '#ffeda0', '#fed976', '#feb24c', '#fd8d3c', '#fc4e2a', '#e31a1c', '#bd0026', '#800026'],
    YlOrBr3: ['#fff7bc', '#fec44f', '#d95f0e'],
    YlOrBr4: ['#ffffd4', '#fed98e', '#fe9929', '#cc4c02'],
    YlOrBr5: ['#ffffd4', '#fed98e', '#fe9929', '#d95f0e', '#993404'],
    YlOrBr6: ['#ffffd4', '#fee391', '#fec44f', '#fe9929', '#d95f0e', '#993404'],
    YlOrBr7: ['#ffffd4', '#fee391', '#fec44f', '#fe9929', '#ec7014', '#cc4c02', '#8c2d04'],
    YlOrBr8: ['#ffffe5', '#fff7bc', '#fee391', '#fec44f', '#fe9929', '#ec7014', '#cc4c02', '#8c2d04'],
    YlOrBr9: ['#ffffe5', '#fff7bc', '#fee391', '#fec44f', '#fe9929', '#ec7014', '#cc4c02', '#993404', '#662506'],
    Purples3: ['#efedf5', '#bcbddc', '#756bb1'],
    Purples4: ['#f2f0f7', '#cbc9e2', '#9e9ac8', '#6a51a3'],
    Purples5: ['#f2f0f7', '#cbc9e2', '#9e9ac8', '#756bb1', '#54278f'],
    Purples6: ['#f2f0f7', '#dadaeb', '#bcbddc', '#9e9ac8', '#756bb1', '#54278f'],
    Purples7: ['#f2f0f7', '#dadaeb', '#bcbddc', '#9e9ac8', '#807dba', '#6a51a3', '#4a1486'],
    Purples8: ['#fcfbfd', '#efedf5', '#dadaeb', '#bcbddc', '#9e9ac8', '#807dba', '#6a51a3', '#4a1486'],
    Purples9: ['#fcfbfd', '#efedf5', '#dadaeb', '#bcbddc', '#9e9ac8', '#807dba', '#6a51a3', '#54278f', '#3f007d'],
    Blues3: ['#deebf7', '#9ecae1', '#3182bd'],
    Blues4: ['#eff3ff', '#bdd7e7', '#6baed6', '#2171b5'],
    Blues5: ['#eff3ff', '#bdd7e7', '#6baed6', '#3182bd', '#08519c'],
    Blues6: ['#eff3ff', '#c6dbef', '#9ecae1', '#6baed6', '#3182bd', '#08519c'],
    Blues7: ['#eff3ff', '#c6dbef', '#9ecae1', '#6baed6', '#4292c6', '#2171b5', '#084594'],
    Blues8: ['#f7fbff', '#deebf7', '#c6dbef', '#9ecae1', '#6baed6', '#4292c6', '#2171b5', '#084594'],
    Blues9: ['#f7fbff', '#deebf7', '#c6dbef', '#9ecae1', '#6baed6', '#4292c6', '#2171b5', '#08519c', '#08306b'],
    Greens3: ['#e5f5e0', '#a1d99b', '#31a354'],
    Greens4: ['#edf8e9', '#bae4b3', '#74c476', '#238b45'],
    Greens5: ['#edf8e9', '#bae4b3', '#74c476', '#31a354', '#006d2c'],
    Greens6: ['#edf8e9', '#c7e9c0', '#a1d99b', '#74c476', '#31a354', '#006d2c'],
    Greens7: ['#edf8e9', '#c7e9c0', '#a1d99b', '#74c476', '#41ab5d', '#238b45', '#005a32'],
    Greens8: ['#f7fcf5', '#e5f5e0', '#c7e9c0', '#a1d99b', '#74c476', '#41ab5d', '#238b45', '#005a32'],
    Greens9: ['#f7fcf5', '#e5f5e0', '#c7e9c0', '#a1d99b', '#74c476', '#41ab5d', '#238b45', '#006d2c', '#00441b'],
    Oranges3: ['#fee6ce', '#fdae6b', '#e6550d'],
    Oranges4: ['#feedde', '#fdbe85', '#fd8d3c', '#d94701'],
    Oranges5: ['#feedde', '#fdbe85', '#fd8d3c', '#e6550d', '#a63603'],
    Oranges6: ['#feedde', '#fdd0a2', '#fdae6b', '#fd8d3c', '#e6550d', '#a63603'],
    Oranges7: ['#feedde', '#fdd0a2', '#fdae6b', '#fd8d3c', '#f16913', '#d94801', '#8c2d04'],
    Oranges8: ['#fff5eb', '#fee6ce', '#fdd0a2', '#fdae6b', '#fd8d3c', '#f16913', '#d94801', '#8c2d04'],
    Oranges9: ['#fff5eb', '#fee6ce', '#fdd0a2', '#fdae6b', '#fd8d3c', '#f16913', '#d94801', '#a63603', '#7f2704'],
    Reds3: ['#fee0d2', '#fc9272', '#de2d26'],
    Reds4: ['#fee5d9', '#fcae91', '#fb6a4a', '#cb181d'],
    Reds5: ['#fee5d9', '#fcae91', '#fb6a4a', '#de2d26', '#a50f15'],
    Reds6: ['#fee5d9', '#fcbba1', '#fc9272', '#fb6a4a', '#de2d26', '#a50f15'],
    Reds7: ['#fee5d9', '#fcbba1', '#fc9272', '#fb6a4a', '#ef3b2c', '#cb181d', '#99000d'],
    Reds8: ['#fff5f0', '#fee0d2', '#fcbba1', '#fc9272', '#fb6a4a', '#ef3b2c', '#cb181d', '#99000d'],
    Reds9: ['#fff5f0', '#fee0d2', '#fcbba1', '#fc9272', '#fb6a4a', '#ef3b2c', '#cb181d', '#a50f15', '#67000d'],
    Greys3: ['#f0f0f0', '#bdbdbd', '#636363'],
    Greys4: ['#f7f7f7', '#cccccc', '#969696', '#525252'],
    Greys5: ['#f7f7f7', '#cccccc', '#969696', '#636363', '#252525'],
    Greys6: ['#f7f7f7', '#d9d9d9', '#bdbdbd', '#969696', '#636363', '#252525'],
    Greys7: ['#f7f7f7', '#d9d9d9', '#bdbdbd', '#969696', '#737373', '#525252', '#252525'],
    Greys8: ['#FFFFFF', '#f0f0f0', '#d9d9d9', '#bdbdbd', '#969696', '#737373', '#525252', '#252525'],
    Greys9: ['#FFFFFF', '#f0f0f0', '#d9d9d9', '#bdbdbd', '#969696', '#737373', '#525252', '#252525', '#000000'],
    // Diverging
    PuOr3: ['#f1a340', '#f7f7f7', '#998ec3'],
    PuOr4: ['#e66101', '#fdb863', '#b2abd2', '#5e3c99'],
    PuOr5: ['#e66101', '#fdb863', '#f7f7f7', '#b2abd2', '#5e3c99'],
    PuOr6: ['#b35806', '#f1a340', '#fee0b6', '#d8daeb', '#998ec3', '#542788'],
    PuOr7: ['#b35806', '#f1a340', '#fee0b6', '#f7f7f7', '#d8daeb', '#998ec3', '#542788'],
    PuOr8: ['#b35806', '#e08214', '#fdb863', '#fee0b6', '#d8daeb', '#b2abd2', '#8073ac', '#542788'],
    PuOr9: ['#b35806', '#e08214', '#fdb863', '#fee0b6', '#f7f7f7', '#d8daeb', '#b2abd2', '#8073ac', '#542788'],
    PuOr10: ['#7f3b08', '#b35806', '#e08214', '#fdb863', '#fee0b6', '#d8daeb', '#b2abd2', '#8073ac', '#542788', '#2d004b'],
    PuOr11: ['#7f3b08', '#b35806', '#e08214', '#fdb863', '#fee0b6', '#f7f7f7', '#d8daeb', '#b2abd2', '#8073ac', '#542788', '#2d004b'],
    BrBG3: ['#d8b365', '#f5f5f5', '#5ab4ac'],
    BrBG4: ['#a6611a', '#dfc27d', '#80cdc1', '#018571'],
    BrBG5: ['#a6611a', '#dfc27d', '#f5f5f5', '#80cdc1', '#018571'],
    BrBG6: ['#8c510a', '#d8b365', '#f6e8c3', '#c7eae5', '#5ab4ac', '#01665e'],
    BrBG7: ['#8c510a', '#d8b365', '#f6e8c3', '#f5f5f5', '#c7eae5', '#5ab4ac', '#01665e'],
    BrBG8: ['#8c510a', '#bf812d', '#dfc27d', '#f6e8c3', '#c7eae5', '#80cdc1', '#35978f', '#01665e'],
    BrBG9: ['#8c510a', '#bf812d', '#dfc27d', '#f6e8c3', '#f5f5f5', '#c7eae5', '#80cdc1', '#35978f', '#01665e'],
    BrBG10: ['#543005', '#8c510a', '#bf812d', '#dfc27d', '#f6e8c3', '#c7eae5', '#80cdc1', '#35978f', '#01665e', '#003c30'],
    BrBG11: ['#543005', '#8c510a', '#bf812d', '#dfc27d', '#f6e8c3', '#f5f5f5', '#c7eae5', '#80cdc1', '#35978f', '#01665e', '#003c30'],
    PRGn3: ['#af8dc3', '#f7f7f7', '#7fbf7b'],
    PRGn4: ['#7b3294', '#c2a5cf', '#a6dba0', '#008837'],
    PRGn5: ['#7b3294', '#c2a5cf', '#f7f7f7', '#a6dba0', '#008837'],
    PRGn6: ['#762a83', '#af8dc3', '#e7d4e8', '#d9f0d3', '#7fbf7b', '#1b7837'],
    PRGn7: ['#762a83', '#af8dc3', '#e7d4e8', '#f7f7f7', '#d9f0d3', '#7fbf7b', '#1b7837'],
    PRGn8: ['#762a83', '#9970ab', '#c2a5cf', '#e7d4e8', '#d9f0d3', '#a6dba0', '#5aae61', '#1b7837'],
    PRGn9: ['#762a83', '#9970ab', '#c2a5cf', '#e7d4e8', '#f7f7f7', '#d9f0d3', '#a6dba0', '#5aae61', '#1b7837'],
    PRGn10: ['#40004b', '#762a83', '#9970ab', '#c2a5cf', '#e7d4e8', '#d9f0d3', '#a6dba0', '#5aae61', '#1b7837', '#00441b'],
    PRGn11: ['#40004b', '#762a83', '#9970ab', '#c2a5cf', '#e7d4e8', '#f7f7f7', '#d9f0d3', '#a6dba0', '#5aae61', '#1b7837', '#00441b'],
    PiYG3: ['#e9a3c9', '#f7f7f7', '#a1d76a'],
    PiYG4: ['#d01c8b', '#f1b6da', '#b8e186', '#4dac26'],
    PiYG5: ['#d01c8b', '#f1b6da', '#f7f7f7', '#b8e186', '#4dac26'],
    PiYG6: ['#c51b7d', '#e9a3c9', '#fde0ef', '#e6f5d0', '#a1d76a', '#4d9221'],
    PiYG7: ['#c51b7d', '#e9a3c9', '#fde0ef', '#f7f7f7', '#e6f5d0', '#a1d76a', '#4d9221'],
    PiYG8: ['#c51b7d', '#de77ae', '#f1b6da', '#fde0ef', '#e6f5d0', '#b8e186', '#7fbc41', '#4d9221'],
    PiYG9: ['#c51b7d', '#de77ae', '#f1b6da', '#fde0ef', '#f7f7f7', '#e6f5d0', '#b8e186', '#7fbc41', '#4d9221'],
    PiYG10: ['#8e0152', '#c51b7d', '#de77ae', '#f1b6da', '#fde0ef', '#e6f5d0', '#b8e186', '#7fbc41', '#4d9221', '#276419'],
    PiYG11: ['#8e0152', '#c51b7d', '#de77ae', '#f1b6da', '#fde0ef', '#f7f7f7', '#e6f5d0', '#b8e186', '#7fbc41', '#4d9221', '#276419'],
    RdBu3: ['#ef8a62', '#f7f7f7', '#67a9cf'],
    RdBu4: ['#ca0020', '#f4a582', '#92c5de', '#0571b0'],
    RdBu5: ['#ca0020', '#f4a582', '#f7f7f7', '#92c5de', '#0571b0'],
    RdBu6: ['#b2182b', '#ef8a62', '#fddbc7', '#d1e5f0', '#67a9cf', '#2166ac'],
    RdBu7: ['#b2182b', '#ef8a62', '#fddbc7', '#f7f7f7', '#d1e5f0', '#67a9cf', '#2166ac'],
    RdBu8: ['#b2182b', '#d6604d', '#f4a582', '#fddbc7', '#d1e5f0', '#92c5de', '#4393c3', '#2166ac'],
    RdBu9: ['#b2182b', '#d6604d', '#f4a582', '#fddbc7', '#f7f7f7', '#d1e5f0', '#92c5de', '#4393c3', '#2166ac'],
    RdBu10: ['#67001f', '#b2182b', '#d6604d', '#f4a582', '#fddbc7', '#d1e5f0', '#92c5de', '#4393c3', '#2166ac', '#053061'],
    RdBu11: ['#67001f', '#b2182b', '#d6604d', '#f4a582', '#fddbc7', '#f7f7f7', '#d1e5f0', '#92c5de', '#4393c3', '#2166ac', '#053061'],
    RdGy3: ['#ef8a62', '#FFFFFF', '#999999'],
    RdGy4: ['#ca0020', '#f4a582', '#bababa', '#404040'],
    RdGy5: ['#ca0020', '#f4a582', '#FFFFFF', '#bababa', '#404040'],
    RdGy6: ['#b2182b', '#ef8a62', '#fddbc7', '#e0e0e0', '#999999', '#4d4d4d'],
    RdGy7: ['#b2182b', '#ef8a62', '#fddbc7', '#FFFFFF', '#e0e0e0', '#999999', '#4d4d4d'],
    RdGy8: ['#b2182b', '#d6604d', '#f4a582', '#fddbc7', '#e0e0e0', '#bababa', '#878787', '#4d4d4d'],
    RdGy9: ['#b2182b', '#d6604d', '#f4a582', '#fddbc7', '#FFFFFF', '#e0e0e0', '#bababa', '#878787', '#4d4d4d'],
    RdGy10: ['#67001f', '#b2182b', '#d6604d', '#f4a582', '#fddbc7', '#e0e0e0', '#bababa', '#878787', '#4d4d4d', '#1a1a1a'],
    RdGy11: ['#67001f', '#b2182b', '#d6604d', '#f4a582', '#fddbc7', '#FFFFFF', '#e0e0e0', '#bababa', '#878787', '#4d4d4d', '#1a1a1a'],
    RdYlBu3: ['#fc8d59', '#ffffbf', '#91bfdb'],
    RdYlBu4: ['#d7191c', '#fdae61', '#abd9e9', '#2c7bb6'],
    RdYlBu5: ['#d7191c', '#fdae61', '#ffffbf', '#abd9e9', '#2c7bb6'],
    RdYlBu6: ['#d73027', '#fc8d59', '#fee090', '#e0f3f8', '#91bfdb', '#4575b4'],
    RdYlBu7: ['#d73027', '#fc8d59', '#fee090', '#ffffbf', '#e0f3f8', '#91bfdb', '#4575b4'],
    RdYlBu8: ['#d73027', '#f46d43', '#fdae61', '#fee090', '#e0f3f8', '#abd9e9', '#74add1', '#4575b4'],
    RdYlBu9: ['#d73027', '#f46d43', '#fdae61', '#fee090', '#ffffbf', '#e0f3f8', '#abd9e9', '#74add1', '#4575b4'],
    RdYlBu10: ['#a50026', '#d73027', '#f46d43', '#fdae61', '#fee090', '#e0f3f8', '#abd9e9', '#74add1', '#4575b4', '#313695'],
    RdYlBu11: ['#a50026', '#d73027', '#f46d43', '#fdae61', '#fee090', '#ffffbf', '#e0f3f8', '#abd9e9', '#74add1', '#4575b4', '#313695'],
    Spectral3: ['#fc8d59', '#ffffbf', '#99d594'],
    Spectral4: ['#d7191c', '#fdae61', '#abdda4', '#2b83ba'],
    Spectral5: ['#d7191c', '#fdae61', '#ffffbf', '#abdda4', '#2b83ba'],
    Spectral6: ['#d53e4f', '#fc8d59', '#fee08b', '#e6f598', '#99d594', '#3288bd'],
    Spectral7: ['#d53e4f', '#fc8d59', '#fee08b', '#ffffbf', '#e6f598', '#99d594', '#3288bd'],
    Spectral8: ['#d53e4f', '#f46d43', '#fdae61', '#fee08b', '#e6f598', '#abdda4', '#66c2a5', '#3288bd'],
    Spectral9: ['#d53e4f', '#f46d43', '#fdae61', '#fee08b', '#ffffbf', '#e6f598', '#abdda4', '#66c2a5', '#3288bd'],
    Spectral10: ['#9e0142', '#d53e4f', '#f46d43', '#fdae61', '#fee08b', '#e6f598', '#abdda4', '#66c2a5', '#3288bd', '#5e4fa2'],
    Spectral11: ['#9e0142', '#d53e4f', '#f46d43', '#fdae61', '#fee08b', '#ffffbf', '#e6f598', '#abdda4', '#66c2a5', '#3288bd', '#5e4fa2'],
    RdYlGn3: ['#fc8d59', '#ffffbf', '#91cf60'],
    RdYlGn4: ['#d7191c', '#fdae61', '#a6d96a', '#1a9641'],
    RdYlGn5: ['#d7191c', '#fdae61', '#ffffbf', '#a6d96a', '#1a9641'],
    RdYlGn6: ['#d73027', '#fc8d59', '#fee08b', '#d9ef8b', '#91cf60', '#1a9850'],
    RdYlGn7: ['#d73027', '#fc8d59', '#fee08b', '#ffffbf', '#d9ef8b', '#91cf60', '#1a9850'],
    RdYlGn8: ['#d73027', '#f46d43', '#fdae61', '#fee08b', '#d9ef8b', '#a6d96a', '#66bd63', '#1a9850'],
    RdYlGn9: ['#d73027', '#f46d43', '#fdae61', '#fee08b', '#ffffbf', '#d9ef8b', '#a6d96a', '#66bd63', '#1a9850'],
    RdYlGn10: ['#a50026', '#d73027', '#f46d43', '#fdae61', '#fee08b', '#d9ef8b', '#a6d96a', '#66bd63', '#1a9850', '#006837'],
    RdYlGn11: ['#a50026', '#d73027', '#f46d43', '#fdae61', '#fee08b', '#ffffbf', '#d9ef8b', '#a6d96a', '#66bd63', '#1a9850', '#006837'],
    // Qualitative
    Accent3: ['#7fc97f', '#beaed4', '#fdc086'],
    Accent4: ['#7fc97f', '#beaed4', '#fdc086', '#ffff99'],
    Accent5: ['#7fc97f', '#beaed4', '#fdc086', '#ffff99', '#386cb0'],
    Accent6: ['#7fc97f', '#beaed4', '#fdc086', '#ffff99', '#386cb0', '#f0027f'],
    Accent7: ['#7fc97f', '#beaed4', '#fdc086', '#ffff99', '#386cb0', '#f0027f', '#bf5b17'],
    Accent8: ['#7fc97f', '#beaed4', '#fdc086', '#ffff99', '#386cb0', '#f0027f', '#bf5b17', '#666666'],
    DarkTwo3: ['#1b9e77', '#d95f02', '#7570b3'],
    DarkTwo4: ['#1b9e77', '#d95f02', '#7570b3', '#e7298a'],
    DarkTwo5: ['#1b9e77', '#d95f02', '#7570b3', '#e7298a', '#66a61e'],
    DarkTwo6: ['#1b9e77', '#d95f02', '#7570b3', '#e7298a', '#66a61e', '#e6ab02'],
    DarkTwo7: ['#1b9e77', '#d95f02', '#7570b3', '#e7298a', '#66a61e', '#e6ab02', '#a6761d'],
    DarkTwo8: ['#1b9e77', '#d95f02', '#7570b3', '#e7298a', '#66a61e', '#e6ab02', '#a6761d', '#666666'],
    Paired3: ['#a6cee3', '#1f78b4', '#b2df8a'],
    Paired4: ['#a6cee3', '#1f78b4', '#b2df8a', '#33a02c'],
    Paired5: ['#a6cee3', '#1f78b4', '#b2df8a', '#33a02c', '#fb9a99'],
    Paired6: ['#a6cee3', '#1f78b4', '#b2df8a', '#33a02c', '#fb9a99', '#e31a1c'],
    Paired7: ['#a6cee3', '#1f78b4', '#b2df8a', '#33a02c', '#fb9a99', '#e31a1c', '#fdbf6f'],
    Paired8: ['#a6cee3', '#1f78b4', '#b2df8a', '#33a02c', '#fb9a99', '#e31a1c', '#fdbf6f', '#ff7f00'],
    Paired9: ['#a6cee3', '#1f78b4', '#b2df8a', '#33a02c', '#fb9a99', '#e31a1c', '#fdbf6f', '#ff7f00', '#cab2d6'],
    Paired10: ['#a6cee3', '#1f78b4', '#b2df8a', '#33a02c', '#fb9a99', '#e31a1c', '#fdbf6f', '#ff7f00', '#cab2d6', '#6a3d9a'],
    Paired11: ['#a6cee3', '#1f78b4', '#b2df8a', '#33a02c', '#fb9a99', '#e31a1c', '#fdbf6f', '#ff7f00', '#cab2d6', '#6a3d9a', '#ffff99'],
    Paired12: ['#a6cee3', '#1f78b4', '#b2df8a', '#33a02c', '#fb9a99', '#e31a1c', '#fdbf6f', '#ff7f00', '#cab2d6', '#6a3d9a', '#ffff99', '#b15928'],
    PastelOne3: ['#fbb4ae', '#b3cde3', '#ccebc5'],
    PastelOne4: ['#fbb4ae', '#b3cde3', '#ccebc5', '#decbe4'],
    PastelOne5: ['#fbb4ae', '#b3cde3', '#ccebc5', '#decbe4', '#fed9a6'],
    PastelOne6: ['#fbb4ae', '#b3cde3', '#ccebc5', '#decbe4', '#fed9a6', '#ffffcc'],
    PastelOne7: ['#fbb4ae', '#b3cde3', '#ccebc5', '#decbe4', '#fed9a6', '#ffffcc', '#e5d8bd'],
    PastelOne8: ['#fbb4ae', '#b3cde3', '#ccebc5', '#decbe4', '#fed9a6', '#ffffcc', '#e5d8bd', '#fddaec'],
    PastelOne9: ['#fbb4ae', '#b3cde3', '#ccebc5', '#decbe4', '#fed9a6', '#ffffcc', '#e5d8bd', '#fddaec', '#f2f2f2'],
    PastelTwo3: ['#b3e2cd', '#fdcdac', '#cbd5e8'],
    PastelTwo4: ['#b3e2cd', '#fdcdac', '#cbd5e8', '#f4cae4'],
    PastelTwo5: ['#b3e2cd', '#fdcdac', '#cbd5e8', '#f4cae4', '#e6f5c9'],
    PastelTwo6: ['#b3e2cd', '#fdcdac', '#cbd5e8', '#f4cae4', '#e6f5c9', '#fff2ae'],
    PastelTwo7: ['#b3e2cd', '#fdcdac', '#cbd5e8', '#f4cae4', '#e6f5c9', '#fff2ae', '#f1e2cc'],
    PastelTwo8: ['#b3e2cd', '#fdcdac', '#cbd5e8', '#f4cae4', '#e6f5c9', '#fff2ae', '#f1e2cc', '#cccccc'],
    SetOne3: ['#e41a1c', '#377eb8', '#4daf4a'],
    SetOne4: ['#e41a1c', '#377eb8', '#4daf4a', '#984ea3'],
    SetOne5: ['#e41a1c', '#377eb8', '#4daf4a', '#984ea3', '#ff7f00'],
    SetOne6: ['#e41a1c', '#377eb8', '#4daf4a', '#984ea3', '#ff7f00', '#ffff33'],
    SetOne7: ['#e41a1c', '#377eb8', '#4daf4a', '#984ea3', '#ff7f00', '#ffff33', '#a65628'],
    SetOne8: ['#e41a1c', '#377eb8', '#4daf4a', '#984ea3', '#ff7f00', '#ffff33', '#a65628', '#f781bf'],
    SetOne9: ['#e41a1c', '#377eb8', '#4daf4a', '#984ea3', '#ff7f00', '#ffff33', '#a65628', '#f781bf', '#999999'],
    SetTwo3: ['#66c2a5', '#fc8d62', '#8da0cb'],
    SetTwo4: ['#66c2a5', '#fc8d62', '#8da0cb', '#e78ac3'],
    SetTwo5: ['#66c2a5', '#fc8d62', '#8da0cb', '#e78ac3', '#a6d854'],
    SetTwo6: ['#66c2a5', '#fc8d62', '#8da0cb', '#e78ac3', '#a6d854', '#ffd92f'],
    SetTwo7: ['#66c2a5', '#fc8d62', '#8da0cb', '#e78ac3', '#a6d854', '#ffd92f', '#e5c494'],
    SetTwo8: ['#66c2a5', '#fc8d62', '#8da0cb', '#e78ac3', '#a6d854', '#ffd92f', '#e5c494', '#b3b3b3'],
    SetThree3: ['#8dd3c7', '#ffffb3', '#bebada'],
    SetThree4: ['#8dd3c7', '#ffffb3', '#bebada', '#fb8072'],
    SetThree5: ['#8dd3c7', '#ffffb3', '#bebada', '#fb8072', '#80b1d3'],
    SetThree6: ['#8dd3c7', '#ffffb3', '#bebada', '#fb8072', '#80b1d3', '#fdb462'],
    SetThree7: ['#8dd3c7', '#ffffb3', '#bebada', '#fb8072', '#80b1d3', '#fdb462', '#b3de69'],
    SetThree8: ['#8dd3c7', '#ffffb3', '#bebada', '#fb8072', '#80b1d3', '#fdb462', '#b3de69', '#fccde5'],
    SetThree9: ['#8dd3c7', '#ffffb3', '#bebada', '#fb8072', '#80b1d3', '#fdb462', '#b3de69', '#fccde5', '#d9d9d9'],
    SetThree10: ['#8dd3c7', '#ffffb3', '#bebada', '#fb8072', '#80b1d3', '#fdb462', '#b3de69', '#fccde5', '#d9d9d9', '#bc80bd'],
    SetThree11: ['#8dd3c7', '#ffffb3', '#bebada', '#fb8072', '#80b1d3', '#fdb462', '#b3de69', '#fccde5', '#d9d9d9', '#bc80bd', '#ccebc5'],
    SetThree12: ['#8dd3c7', '#ffffb3', '#bebada', '#fb8072', '#80b1d3', '#fdb462', '#b3de69', '#fccde5', '#d9d9d9', '#bc80bd', '#ccebc5', '#ffed6f'],
};

var office = {
    Adjacency6: ['#a9a57c', '#9cbebd', '#d2cb6c', '#95a39d', '#c89f5d', '#b1a089'],
    Advantage6: ['#663366', '#330f42', '#666699', '#999966', '#f7901e', '#a3a101'],
    Angles6: ['#797b7e', '#f96a1b', '#08a1d9', '#7c984a', '#c2ad8d', '#506e94'],
    Apex6: ['#ceb966', '#9cb084', '#6bb1c9', '#6585cf', '#7e6bc9', '#a379bb'],
    Apothecary6: ['#93a299', '#cf543f', '#b5ae53', '#848058', '#e8b54d', '#786c71'],
    Aspect6: ['#f07f09', '#9f2936', '#1b587c', '#4e8542', '#604878', '#c19859'],
    Atlas6: ['#f81b02', '#fc7715', '#afbf41', '#50c49f', '#3b95c4', '#b560d4'],
    Austin6: ['#94c600', '#71685a', '#ff6700', '#909465', '#956b43', '#fea022'],
    Badge6: ['#f8b323', '#656a59', '#46b2b5', '#8caa7e', '#d36f68', '#826276'],
    Banded6: ['#ffc000', '#a5d028', '#08cc78', '#f24099', '#828288', '#f56617'],
    Basis6: ['#f09415', '#c1b56b', '#4baf73', '#5aa6c0', '#d17df9', '#fa7e5c'],
    Berlin6: ['#a6b727', '#df5327', '#fe9e00', '#418ab3', '#d7d447', '#818183'],
    BlackTie6: ['#6f6f74', '#a7b789', '#beae98', '#92a9b9', '#9c8265', '#8d6974'],
    Blue6: ['#0f6fc6', '#009dd9', '#0bd0d9', '#10cf9b', '#7cca62', '#a5c249'],
    BlueGreen6: ['#3494ba', '#58b6c0', '#75bda7', '#7a8c8e', '#84acb6', '#2683c6'],
    BlueII6: ['#1cade4', '#2683c6', '#27ced7', '#42ba97', '#3e8853', '#62a39f'],
    BlueRed6: ['#4a66ac', '#629dd1', '#297fd5', '#7f8fa9', '#5aa2ae', '#9d90a0'],
    BlueWarm6: ['#4a66ac', '#629dd1', '#297fd5', '#7f8fa9', '#5aa2ae', '#9d90a0'],
    Breeze6: ['#2c7c9f', '#244a58', '#e2751d', '#ffb400', '#7eb606', '#c00000'],
    Capital6: ['#4b5a60', '#9c5238', '#504539', '#c1ad79', '#667559', '#bad6ad'],
    Celestial6: ['#ac3ec1', '#477bd1', '#46b298', '#90ba4c', '#dd9d31', '#e25247'],
    Circuit6: ['#9acd4c', '#faa93a', '#d35940', '#b258d3', '#63a0cc', '#8ac4a7'],
    Civic6: ['#d16349', '#ccb400', '#8cadae', '#8c7b70', '#8fb08c', '#d19049'],
    Clarity6: ['#93a299', '#ad8f67', '#726056', '#4c5a6a', '#808da0', '#79463d'],
    Codex6: ['#990000', '#efab16', '#78ac35', '#35aca2', '#4083cf', '#0d335e'],
    Composite6: ['#98c723', '#59b0b9', '#deae00', '#b77bb4', '#e0773c', '#a98d63'],
    Concourse6: ['#2da2bf', '#da1f28', '#eb641b', '#39639d', '#474b78', '#7d3c4a'],
    Couture6: ['#9e8e5c', '#a09781', '#85776d', '#aeafa9', '#8d878b', '#6b6149'],
    Crop6: ['#8c8d86', '#e6c069', '#897b61', '#8dab8e', '#77a2bb', '#e28394'],
    Damask6: ['#9ec544', '#50bea3', '#4a9ccc', '#9a66ca', '#c54f71', '#de9c3c'],
    Depth6: ['#41aebd', '#97e9d5', '#a2cf49', '#608f3d', '#f4de3a', '#fcb11c'],
    Dividend6: ['#4d1434', '#903163', '#b2324b', '#969fa7', '#66b1ce', '#40619d'],
    Droplet6: ['#2fa3ee', '#4bcaad', '#86c157', '#d99c3f', '#ce6633', '#a35dd1'],
    Elemental6: ['#629dd1', '#297fd5', '#7f8fa9', '#4a66ac', '#5aa2ae', '#9d90a0'],
    Equity6: ['#d34817', '#9b2d1f', '#a28e6a', '#956251', '#918485', '#855d5d'],
    Essential6: ['#7a7a7a', '#f5c201', '#526db0', '#989aac', '#dc5924', '#b4b392'],
    Excel16: ['#9999ff', '#993366', '#ffffcc', '#ccffff', '#660066', '#ff8080', '#0066cc', '#ccccff', '#000080', '#ff00ff', '#ffff00', '#0000ff', '#800080', '#800000', '#008080', '#0000ff'],
    Executive6: ['#6076b4', '#9c5252', '#e68422', '#846648', '#63891f', '#758085'],
    Exhibit6: ['#3399ff', '#69ffff', '#ccff33', '#3333ff', '#9933ff', '#ff33ff'],
    Expo6: ['#fbc01e', '#efe1a2', '#fa8716', '#be0204', '#640f10', '#7e13e3'],
    Facet6: ['#90c226', '#54a021', '#e6b91e', '#e76618', '#c42f1a', '#918655'],
    Feathered6: ['#606372', '#79a8a4', '#b2ad8f', '#ad8082', '#dec18c', '#92a185'],
    Flow6: ['#0f6fc6', '#009dd9', '#0bd0d9', '#10cf9b', '#7cca62', '#a5c249'],
    Focus6: ['#ffb91d', '#f97817', '#6de304', '#ff0000', '#732bea', '#c913ad'],
    Folio6: ['#294171', '#748cbc', '#8e887c', '#834736', '#5a1705', '#a0a16a'],
    Formal6: ['#907f76', '#a46645', '#cd9c47', '#9a92cd', '#7d639b', '#733678'],
    Forte6: ['#c70f0c', '#dd6b0d', '#faa700', '#93e50d', '#17c7ba', '#0a96e4'],
    Foundry6: ['#72a376', '#b0ccb0', '#a8cdd7', '#c0beaf', '#cec597', '#e8b7b7'],
    Frame6: ['#40bad2', '#fab900', '#90bb23', '#ee7008', '#1ab39f', '#d5393d'],
    Gallery6: ['#b71e42', '#de478e', '#bc72f0', '#795faf', '#586ea6', '#6892a0'],
    Genesis6: ['#80b606', '#e29f1d', '#2397e2', '#35aca2', '#5430bb', '#8d34e0'],
    Grayscale6: ['#dddddd', '#b2b2b2', '#969696', '#808080', '#5f5f5f', '#4d4d4d'],
    Green6: ['#549e39', '#8ab833', '#c0cf3a', '#029676', '#4ab5c4', '#0989b1'],
    GreenYellow6: ['#99cb38', '#63a537', '#37a76f', '#44c1a3', '#4eb3cf', '#51c3f9'],
    Grid6: ['#c66951', '#bf974d', '#928b70', '#87706b', '#94734e', '#6f777d'],
    Habitat6: ['#f8c000', '#f88600', '#f83500', '#8b723d', '#818b3d', '#586215'],
    Hardcover6: ['#873624', '#d6862d', '#d0be40', '#877f6c', '#972109', '#aeb795'],
    Headlines6: ['#439eb7', '#e28b55', '#dcb64d', '#4ca198', '#835b82', '#645135'],
    Horizon6: ['#7e97ad', '#cc8e60', '#7a6a60', '#b4936d', '#67787b', '#9d936f'],
    Infusion6: ['#8c73d0', '#c2e8c4', '#c5a6e8', '#b45ec7', '#9fdafb', '#95c5b0'],
    Inkwell6: ['#860908', '#4a0505', '#7a500a', '#c47810', '#827752', '#b5bb83'],
    Inspiration6: ['#749805', '#bacc82', '#6e9ec2', '#2046a5', '#5039c6', '#7411d0'],
    Integral6: ['#1cade4', '#2683c6', '#27ced7', '#42ba97', '#3e8853', '#62a39f'],
    Ion6: ['#b01513', '#ea6312', '#e6b729', '#6aac90', '#5f9c9d', '#9e5e9b'],
    IonBoardroom6: ['#b31166', '#e33d6f', '#e45f3c', '#e9943a', '#9b6bf2', '#d53dd0'],
    Kilter6: ['#76c5ef', '#fea022', '#ff6700', '#70a525', '#a5d848', '#20768c'],
    Madison6: ['#a1d68b', '#5ec795', '#4dadcf', '#cdb756', '#e29c36', '#8ec0c1'],
    MainEvent6: ['#b80e0f', '#a6987d', '#7f9a71', '#64969f', '#9b75b2', '#80737a'],
    Marquee6: ['#418ab3', '#a6b727', '#f69200', '#838383', '#fec306', '#df5327'],
    Median6: ['#94b6d2', '#dd8047', '#a5ab81', '#d8b25c', '#7ba79d', '#968c8c'],
    Mesh6: ['#6f6f6f', '#bfbfa5', '#dcd084', '#e7bf5f', '#e9a039', '#cf7133'],
    Metail6: ['#6283ad', '#324966', '#5b9ea4', '#1d5b57', '#1b4430', '#2f3c35'],
    Metro6: ['#7fd13b', '#ea157a', '#feb80a', '#00addc', '#738ac8', '#1ab39f'],
    Metropolitan6: ['#50b4c8', '#a8b97f', '#9b9256', '#657689', '#7a855d', '#84ac9d'],
    Module6: ['#f0ad00', '#60b5cc', '#e66c7d', '#6bb76d', '#e88651', '#c64847'],
    NewsPrint6: ['#ad0101', '#726056', '#ac956e', '#808da9', '#424e5b', '#730e00'],
    Office6: ['#5b9bd5', '#ed7d31', '#a5a5a5', '#ffc000', '#4472c4', '#70ad47'],
    OfficeClassic6: ['#4f81bd', '#c0504d', '#9bbb59', '#8064a2', '#4bacc6', '#f79646'],
    Opulent6: ['#b83d68', '#ac66bb', '#de6c36', '#f9b639', '#cf6da4', '#fa8d3d'],
    Orange6: ['#e48312', '#bd582c', '#865640', '#9b8357', '#c2bc80', '#94a088'],
    OrangeRed6: ['#d34817', '#9b2d1f', '#a28e6a', '#956251', '#918485', '#855d5d'],
    Orbit6: ['#f2d908', '#9de61e', '#0d8be6', '#c61b1b', '#e26f08', '#8d35d1'],
    Organic6: ['#83992a', '#3c9770', '#44709d', '#a23c33', '#d97828', '#deb340'],
    Oriel6: ['#fe8637', '#7598d9', '#b32c16', '#f5cd2d', '#aebad5', '#777c84'],
    Origin6: ['#727ca3', '#9fb8cd', '#d2da7a', '#fada7a', '#b88472', '#8e736a'],
    Paper6: ['#a5b592', '#f3a447', '#e7bc29', '#d092a7', '#9c85c0', '#809ec2'],
    Parallax6: ['#30acec', '#80c34f', '#e29d3e', '#d64a3b', '#d64787', '#a666e1'],
    Parcel6: ['#f6a21d', '#9bafb5', '#c96731', '#9ca383', '#87795d', '#a0988c'],
    Perception6: ['#a2c816', '#e07602', '#e4c402', '#7dc1ef', '#21449b', '#a2b170'],
    Perspective6: ['#838d9b', '#d2610c', '#80716a', '#94147c', '#5d5ad2', '#6f6c7d'],
    Pixel6: ['#ff7f01', '#f1b015', '#fbec85', '#d2c2f1', '#da5af4', '#9d09d1'],
    Plaza6: ['#990000', '#580101', '#e94a00', '#eb8f00', '#a4a4a4', '#666666'],
    Precedent6: ['#993232', '#9b6c34', '#736c5d', '#c9972b', '#c95f2b', '#8f7a05'],
    Pushpin6: ['#fda023', '#aa2b1e', '#71685c', '#64a73b', '#eb5605', '#b9ca1a'],
    Quotable6: ['#00c6bb', '#6feba0', '#b6df5e', '#efb251', '#ef755f', '#ed515c'],
    Red6: ['#a5300f', '#d55816', '#e19825', '#b19c7d', '#7f5f52', '#b27d49'],
    RedOrange6: ['#e84c22', '#ffbd47', '#b64926', '#ff8427', '#cc9900', '#b22600'],
    RedViolet6: ['#e32d91', '#c830cc', '#4ea6dc', '#4775e7', '#8971e1', '#d54773'],
    Retrospect6: ['#e48312', '#bd582c', '#865640', '#9b8357', '#c2bc80', '#94a088'],
    Revolution6: ['#0c5986', '#ddf53d', '#508709', '#bf5e00', '#9c0001', '#660075'],
    Saddle6: ['#c6b178', '#9c5b14', '#71b2bc', '#78aa5d', '#867099', '#4c6f75'],
    Savon6: ['#1cade4', '#2683c6', '#27ced7', '#42ba97', '#3e8853', '#62a39f'],
    Sketchbook6: ['#a63212', '#e68230', '#9bb05e', '#6b9bc7', '#4e66b2', '#8976ac'],
    Sky6: ['#073779', '#8fd9fb', '#ffcc00', '#eb6615', '#c76402', '#b523b4'],
    Slate6: ['#bc451b', '#d3ba68', '#bb8640', '#ad9277', '#a55a43', '#ad9d7b'],
    Slice6: ['#052f61', '#a50e82', '#14967c', '#6a9e1f', '#e87d37', '#c62324'],
    Slipstream6: ['#4e67c8', '#5eccf3', '#a7ea52', '#5dceaf', '#ff8021', '#f14124'],
    SOHO6: ['#61625e', '#964d2c', '#66553e', '#848058', '#afa14b', '#ad7d4d'],
    Solstice6: ['#3891a7', '#feb80a', '#c32d2e', '#84aa33', '#964305', '#475a8d'],
    Spectrum6: ['#990000', '#ff6600', '#ffba00', '#99cc00', '#528a02', '#333333'],
    Story6: ['#1d86cd', '#732e9a', '#b50b1b', '#e8950e', '#55992b', '#2c9c89'],
    Studio6: ['#f7901e', '#fec60b', '#9fe62f', '#4ea5d1', '#1c4596', '#542d90'],
    Summer6: ['#51a6c2', '#51c2a9', '#7ec251', '#e1dc53', '#b54721', '#a16bb1'],
    Technic6: ['#6ea0b0', '#ccaf0a', '#8d89a4', '#748560', '#9e9273', '#7e848d'],
    Thatch6: ['#759aa5', '#cfc60d', '#99987f', '#90ac97', '#ffad1c', '#b9ab6f'],
    Tradition6: ['#6b4a0b', '#790a14', '#908342', '#423e5c', '#641345', '#748a2f'],
    Travelogue6: ['#b74d21', '#a32323', '#4576a3', '#615d9a', '#67924b', '#bf7b1b'],
    Trek6: ['#f0a22e', '#a5644e', '#b58b80', '#c3986d', '#a19574', '#c17529'],
    Twilight6: ['#e8bc4a', '#83c1c6', '#e78d35', '#909ce1', '#839c41', '#cc5439'],
    Urban6: ['#53548a', '#438086', '#a04da3', '#c4652d', '#8b5d3d', '#5c92b5'],
    UrbanPop6: ['#86ce24', '#00a2e6', '#fac810', '#7d8f8c', '#d06b20', '#958b8b'],
    VaporTrail6: ['#df2e28', '#fe801a', '#e9bf35', '#81bb42', '#32c7a9', '#4a9bdc'],
    Venture6: ['#9eb060', '#d09a08', '#f2ec86', '#824f1c', '#511818', '#553876'],
    Verve6: ['#ff388c', '#e40059', '#9c007f', '#68007f', '#005bd3', '#00349e'],
    View6: ['#6f6f74', '#92a9b9', '#a7b789', '#b9a489', '#8d6374', '#9b7362'],
    Violet6: ['#ad84c6', '#8784c7', '#5d739a', '#6997af', '#84acb6', '#6f8183'],
    VioletII6: ['#92278f', '#9b57d3', '#755dd9', '#665eb8', '#45a5ed', '#5982db'],
    Waveform6: ['#31b6fd', '#4584d3', '#5bd078', '#a5d028', '#f5c040', '#05e0db'],
    Wisp6: ['#a53010', '#de7e18', '#9f8351', '#728653', '#92aa4c', '#6aac91'],
    WoodType6: ['#d34817', '#9b2d1f', '#a28e6a', '#956251', '#918485', '#855d5d'],
    Yellow6: ['#ffca08', '#f8931d', '#ce8d3e', '#ec7016', '#e64823', '#9c6a6a'],
    YellowOrange6: ['#f0a22e', '#a5644e', '#b58b80', '#c3986d', '#a19574', '#c17529'],
};

var tableau = {
    // New
    Tableau10: ['#4E79A7', '#F28E2B', '#E15759', '#76B7B2', '#59A14F', '#EDC948', '#B07AA1', '#FF9DA7', '#9C755F', '#BAB0AC'],
    Tableau20: ['#4E79A7', '#A0CBE8', '#F28E2B', '#FFBE7D', '#59A14F', '#8CD17D', '#B6992D', '#F1CE63', '#499894', '#86BCB6', '#E15759', '#FF9D9A', '#79706E', '#BAB0AC', '#D37295', '#FABFD2', '#B07AA1', '#D4A6C8', '#9D7660', '#D7B5A6'],
    ColorBlind10: ['#1170aa', '#fc7d0b', '#a3acb9', '#57606c', '#5fa2ce', '#c85200', '#7b848f', '#a3cce9', '#ffbc79', '#c8d0d9'],
    SeattleGrays5: ['#767f8b', '#b3b7b8', '#5c6068', '#d3d3d3', '#989ca3'],
    Traffic9: ['#b60a1c', '#e39802', '#309143', '#e03531', '#f0bd27', '#51b364', '#ff684c', '#ffda66', '#8ace7e'],
    MillerStone11: ['#4f6980', '#849db1', '#a2ceaa', '#638b66', '#bfbb60', '#f47942', '#fbb04e', '#b66353', '#d7ce9f', '#b9aa97', '#7e756d'],
    SuperfishelStone10: ['#6388b4', '#ffae34', '#ef6f6a', '#8cc2ca', '#55ad89', '#c3bc3f', '#bb7693', '#baa094', '#a9b5ae', '#767676'],
    NurielStone9: ['#8175aa', '#6fb899', '#31a1b3', '#ccb22b', '#a39fc9', '#94d0c0', '#959c9e', '#027b8e', '#9f8f12'],
    JewelBright9: ['#eb1e2c', '#fd6f30', '#f9a729', '#f9d23c', '#5fbb68', '#64cdcc', '#91dcea', '#a4a4d5', '#bbc9e5'],
    Summer8: ['#bfb202', '#b9ca5d', '#cf3e53', '#f1788d', '#00a2b3', '#97cfd0', '#f3a546', '#f7c480'],
    Winter10: ['#90728f', '#b9a0b4', '#9d983d', '#cecb76', '#e15759', '#ff9888', '#6b6b6b', '#bab2ae', '#aa8780', '#dab6af'],
    GreenOrangeTeal12: ['#4e9f50', '#87d180', '#ef8a0c', '#fcc66d', '#3ca8bc', '#98d9e4', '#94a323', '#c3ce3d', '#a08400', '#f7d42a', '#26897e', '#8dbfa8'],
    RedBlueBrown12: ['#466f9d', '#91b3d7', '#ed444a', '#feb5a2', '#9d7660', '#d7b5a6', '#3896c4', '#a0d4ee', '#ba7e45', '#39b87f', '#c8133b', '#ea8783'],
    PurplePinkGray12: ['#8074a8', '#c6c1f0', '#c46487', '#ffbed1', '#9c9290', '#c5bfbe', '#9b93c9', '#ddb5d5', '#7c7270', '#f498b6', '#b173a0', '#c799bc'],
    HueCircle19: ['#1ba3c6', '#2cb5c0', '#30bcad', '#21B087', '#33a65c', '#57a337', '#a2b627', '#d5bb21', '#f8b620', '#f89217', '#f06719', '#e03426', '#f64971', '#fc719e', '#eb73b3', '#ce69be', '#a26dc2', '#7873c0', '#4f7cba'],
    OrangeBlue7: ['#9e3d22', '#d45b21', '#f69035', '#d9d5c9', '#77acd3', '#4f81af', '#2b5c8a'],
    RedGreen7: ['#a3123a', '#e33f43', '#f8816b', '#ced7c3', '#73ba67', '#44914e', '#24693d'],
    GreenBlue7: ['#24693d', '#45934d', '#75bc69', '#c9dad2', '#77a9cf', '#4e7fab', '#2a5783'],
    RedBlue7: ['#a90c38', '#e03b42', '#f87f69', '#dfd4d1', '#7eaed3', '#5383af', '#2e5a87'],
    RedBlack7: ['#ae123a', '#e33e43', '#f8816b', '#d9d9d9', '#a0a7a8', '#707c83', '#49525e'],
    GoldPurple7: ['#ad9024', '#c1a33b', '#d4b95e', '#e3d8cf', '#d4a3c3', '#c189b0', '#ac7299'],
    RedGreenGold7: ['#be2a3e', '#e25f48', '#f88f4d', '#f4d166', '#90b960', '#4b9b5f', '#22763f'],
    SunsetSunrise7: ['#33608c', '#9768a5', '#e7718a', '#f6ba57', '#ed7846', '#d54c45', '#b81840'],
    OrangeBlueWhite7: ['#9e3d22', '#e36621', '#fcad52', '#FFFFFF', '#95c5e1', '#5b8fbc', '#2b5c8a'],
    RedGreenWhite7: ['#ae123a', '#ee574d', '#fdac9e', '#FFFFFF', '#91d183', '#539e52', '#24693d'],
    GreenBlueWhite7: ['#24693d', '#529c51', '#8fd180', '#FFFFFF', '#95c1dd', '#598ab5', '#2a5783'],
    RedBlueWhite7: ['#a90c38', '#ec534b', '#feaa9a', '#FFFFFF', '#9ac4e1', '#5c8db8', '#2e5a87'],
    RedBlackWhite7: ['#ae123a', '#ee574d', '#fdac9d', '#FFFFFF', '#bdc0bf', '#7d888d', '#49525e'],
    OrangeBlueLight7: ['#ffcc9e', '#f9d4b6', '#f0dccd', '#e5e5e5', '#dae1ea', '#cfdcef', '#c4d8f3'],
    Temperature7: ['#529985', '#6c9e6e', '#99b059', '#dbcf47', '#ebc24b', '#e3a14f', '#c26b51'],
    BlueGreen7: ['#feffd9', '#f2fabf', '#dff3b2', '#c4eab1', '#94d6b7', '#69c5be', '#41b7c4'],
    BlueLight7: ['#e5e5e5', '#e0e3e8', '#dbe1ea', '#d5dfec', '#d0dcef', '#cadaf1', '#c4d8f3'],
    OrangeLight7: ['#e5e5e5', '#ebe1d9', '#f0ddcd', '#f5d9c2', '#f9d4b6', '#fdd0aa', '#ffcc9e'],
    Blue20: ['#b9ddf1', '#afd6ed', '#a5cfe9', '#9bc7e4', '#92c0df', '#89b8da', '#80b0d5', '#79aacf', '#72a3c9', '#6a9bc3', '#6394be', '#5b8cb8', '#5485b2', '#4e7fac', '#4878a6', '#437a9f', '#3d6a98', '#376491', '#305d8a', '#2a5783'],
    Orange20: ['#ffc685', '#fcbe75', '#f9b665', '#f7ae54', '#f5a645', '#f59c3c', '#f49234', '#f2882d', '#f07e27', '#ee7422', '#e96b20', '#e36420', '#db5e20', '#d25921', '#ca5422', '#c14f22', '#b84b23', '#af4623', '#a64122', '#9e3d22'],
    Green20: ['#b3e0a6', '#a5db96', '#98d687', '#8ed07f', '#85ca77', '#7dc370', '#75bc69', '#6eb663', '#67af5c', '#61a956', '#59a253', '#519c51', '#49964f', '#428f4d', '#398949', '#308344', '#2b7c40', '#27763d', '#256f3d', '#24693d'],
    Red20: ['#ffbeb2', '#feb4a6', '#fdab9b', '#fca290', '#fb9984', '#fa8f79', '#f9856e', '#f77b66', '#f5715d', '#f36754', '#f05c4d', '#ec5049', '#e74545', '#e13b42', '#da323f', '#d3293d', '#ca223c', '#c11a3b', '#b8163a', '#ae123a'],
    Purple20: ['#eec9e5', '#eac1df', '#e6b9d9', '#e0b2d2', '#daabcb', '#d5a4c4', '#cf9dbe', '#ca96b8', '#c48fb2', '#be89ac', '#b882a6', '#b27ba1', '#aa759d', '#a27099', '#9a6a96', '#926591', '#8c5f86', '#865986', '#81537f', '#7c4d79'],
    Brown20: ['#eedbbd', '#ecd2ad', '#ebc994', '#eac085', '#e8b777', '#e5ae6c', '#e2a562', '#de9d5a', '#d99455', '#d38c54', '#ce8451', '#c9784d', '#c47247', '#c16941', '#bd6036', '#b85636', '#b34d34', '#ad4433', '#a63d32', '#9f3632'],
    Gray20: ['#d5d5d5', '#cdcecd', '#c5c7c6', '#bcbfbe', '#b4b7b7', '#acb0b1', '#a4a9ab', '#9ca3a4', '#939c9e', '#8b9598', '#848e93', '#7c878d', '#758087', '#6e7a81', '#67737c', '#616c77', '#5b6570', '#555f6a', '#4f5864', '#49525e'],
    GrayWarm20: ['#dcd4d0', '#d4ccc8', '#cdc4c0', '#c5bdb9', '#beb6b2', '#b7afab', '#b0a7a4', '#a9a09d', '#a29996', '#9b938f', '#948c88', '#8d8481', '#867e7b', '#807774', '#79706e', '#736967', '#6c6260', '#665c51', '#5f5654', '#59504e'],
    BlueTeal20: ['#bce4d8', '#aedcd5', '#a1d5d2', '#95cecf', '#89c8cc', '#7ec1ca', '#72bac6', '#66b2c2', '#59acbe', '#4ba5ba', '#419eb6', '#3b96b2', '#358ead', '#3586a7', '#347ea1', '#32779b', '#316f96', '#2f6790', '#2d608a', '#2c5985'],
    OrangeGold20: ['#f4d166', '#f6c760', '#f8bc58', '#f8b252', '#f7a84a', '#f69e41', '#f49538', '#f38b2f', '#f28026', '#f0751e', '#eb6c1c', '#e4641e', '#de5d1f', '#d75521', '#cf4f22', '#c64a22', '#bc4623', '#b24223', '#a83e24', '#9e3a26'],
    GreenGold20: ['#f4d166', '#e3cd62', '#d3c95f', '#c3c55d', '#b2c25b', '#a3bd5a', '#93b958', '#84b457', '#76af56', '#67a956', '#5aa355', '#4f9e53', '#479751', '#40914f', '#3a8a4d', '#34844a', '#2d7d45', '#257740', '#1c713b', '#146c36'],
    RedGold21: ['#f4d166', '#f5c75f', '#f6bc58', '#f7b254', '#f9a750', '#fa9d4f', '#fa9d4f', '#fb934d', '#f7894b', '#f47f4a', '#f0774a', '#eb6349', '#e66549', '#e15c48', '#dc5447', '#d64c45', '#d04344', '#ca3a42', '#c43141', '#bd273f', '#b71d3e'],
    // Classic
    Classic10: ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728', '#9467bd', '#8c564b', '#e377c2', '#7f7f7f', '#bcbd22', '#17becf'],
    ClassicMedium10: ['#729ece', '#ff9e4a', '#67bf5c', '#ed665d', '#ad8bc9', '#a8786e', '#ed97ca', '#a2a2a2', '#cdcc5d', '#6dccda'],
    ClassicLight10: ['#aec7e8', '#ffbb78', '#98df8a', '#ff9896', '#c5b0d5', '#c49c94', '#f7b6d2', '#c7c7c7', '#dbdb8d', '#9edae5'],
    Classic20: ['#1f77b4', '#aec7e8', '#ff7f0e', '#ffbb78', '#2ca02c', '#98df8a', '#d62728', '#ff9896', '#9467bd', '#c5b0d5', '#8c564b', '#c49c94', '#e377c2', '#f7b6d2', '#7f7f7f', '#c7c7c7', '#bcbd22', '#dbdb8d', '#17becf', '#9edae5'],
    ClassicGray5: ['#60636a', '#a5acaf', '#414451', '#8f8782', '#cfcfcf'],
    ClassicColorBlind10: ['#006ba4', '#ff800e', '#ababab', '#595959', '#5f9ed1', '#c85200', '#898989', '#a2c8ec', '#ffbc79', '#cfcfcf'],
    ClassicTrafficLight9: ['#b10318', '#dba13a', '#309343', '#d82526', '#ffc156', '#69b764', '#f26c64', '#ffdd71', '#9fcd99'],
    ClassicPurpleGray6: ['#7b66d2', '#dc5fbd', '#94917b', '#995688', '#d098ee', '#d7d5c5'],
    ClassicPurpleGray12: ['#7b66d2', '#a699e8', '#dc5fbd', '#ffc0da', '#5f5a41', '#b4b19b', '#995688', '#d898ba', '#ab6ad5', '#d098ee', '#8b7c6e', '#dbd4c5'],
    ClassicGreenOrange6: ['#32a251', '#ff7f0f', '#3cb7cc', '#ffd94a', '#39737c', '#b85a0d'],
    ClassicGreenOrange12: ['#32a251', '#acd98d', '#ff7f0f', '#ffb977', '#3cb7cc', '#98d9e4', '#b85a0d', '#ffd94a', '#39737c', '#86b4a9', '#82853b', '#ccc94d'],
    ClassicBlueRed6: ['#2c69b0', '#f02720', '#ac613c', '#6ba3d6', '#ea6b73', '#e9c39b'],
    ClassicBlueRed12: ['#2c69b0', '#b5c8e2', '#f02720', '#ffb6b0', '#ac613c', '#e9c39b', '#6ba3d6', '#b5dffd', '#ac8763', '#ddc9b4', '#bd0a36', '#f4737a'],
    ClassicCyclic13: ['#1f83b4', '#12a2a8', '#2ca030', '#78a641', '#bcbd22', '#ffbf50', '#ffaa0e', '#ff7f0e', '#d63a3a', '#c7519c', '#ba43b4', '#8a60b0', '#6f63bb'],
    ClassicGreen7: ['#bccfb4', '#94bb83', '#69a761', '#339444', '#27823b', '#1a7232', '#09622a'],
    ClassicGray13: ['#c3c3c3', '#b2b2b2', '#a2a2a2', '#929292', '#838383', '#747474', '#666666', '#585858', '#4b4b4b', '#3f3f3f', '#333333', '#282828', '#1e1e1e'],
    ClassicBlue7: ['#b4d4da', '#7bc8e2', '#67add4', '#3a87b7', '#1c73b1', '#1c5998', '#26456e'],
    ClassicRed9: ['#eac0bd', '#f89a90', '#f57667', '#e35745', '#d8392c', '#cf1719', '#c21417', '#b10c1d', '#9c0824'],
    ClassicOrange7: ['#f0c294', '#fdab67', '#fd8938', '#f06511', '#d74401', '#a33202', '#7b3014'],
    ClassicAreaRed11: ['#f5cac7', '#fbb3ab', '#fd9c8f', '#fe8b7a', '#fd7864', '#f46b55', '#ea5e45', '#e04e35', '#d43e25', '#c92b14', '#bd1100'],
    ClassicAreaGreen11: ['#dbe8b4', '#c3e394', '#acdc7a', '#9ad26d', '#8ac765', '#7abc5f', '#6cae59', '#60a24d', '#569735', '#4a8c1c', '#3c8200'],
    ClassicAreaBrown11: ['#f3e0c2', '#f6d29c', '#f7c577', '#f0b763', '#e4aa63', '#d89c63', '#cc8f63', '#c08262', '#bb7359', '#bb6348', '#bb5137'],
    ClassicRedGreen11: ['#9c0824', '#bd1316', '#d11719', '#df513f', '#fc8375', '#cacaca', '#a2c18f', '#69a761', '#2f8e41', '#1e7735', '#09622a'],
    ClassicRedBlue11: ['#9c0824', '#bd1316', '#d11719', '#df513f', '#fc8375', '#cacaca', '#67add4', '#3a87b7', '#1c73b1', '#1c5998', '#26456e'],
    ClassicRedBlack11: ['#9c0824', '#bd1316', '#d11719', '#df513f', '#fc8375', '#cacaca', '#9b9b9b', '#777777', '#565656', '#383838', '#1e1e1e'],
    ClassicAreaRedGreen21: ['#bd1100', '#c82912', '#d23a21', '#dc4930', '#e6583e', '#ef654d', '#f7705b', '#fd7e6b', '#fe8e7e', '#fca294', '#e9dabe', '#c7e298', '#b1de7f', '#a0d571', '#90cb68', '#82c162', '#75b65d', '#69aa56', '#5ea049', '#559633', '#4a8c1c'],
    ClassicOrangeBlue13: ['#7b3014', '#a33202', '#d74401', '#f06511', '#fd8938', '#fdab67', '#cacaca', '#7bc8e2', '#67add4', '#3a87b7', '#1c73b1', '#1c5998', '#26456e'],
    ClassicGreenBlue11: ['#09622a', '#1e7735', '#2f8e41', '#69a761', '#a2c18f', '#cacaca', '#67add4', '#3a87b7', '#1c73b1', '#1c5998', '#26456e'],
    ClassicRedWhiteGreen11: ['#9c0824', '#b41f27', '#cc312b', '#e86753', '#fcb4a5', '#FFFFFF', '#b9d7b7', '#74af72', '#428f49', '#297839', '#09622a'],
    ClassicRedWhiteBlack11: ['#9c0824', '#b41f27', '#cc312b', '#e86753', '#fcb4a5', '#FFFFFF', '#bfbfbf', '#838383', '#575757', '#393939', '#1e1e1e'],
    ClassicOrangeWhiteBlue11: ['#7b3014', '#a84415', '#d85a13', '#fb8547', '#ffc2a1', '#FFFFFF', '#b7cde2', '#6a9ec5', '#3679a8', '#2e5f8a', '#26456e'],
    ClassicRedWhiteBlackLight10: ['#ffc2c5', '#ffd1d3', '#ffe0e1', '#fff0f0', '#FFFFFF', '#f3f3f3', '#e8e8e8', '#dddddd', '#d1d1d1', '#c6c6c6'],
    ClassicOrangeWhiteBlueLight11: ['#ffcc9e', '#ffd6b1', '#ffe0c5', '#ffead8', '#fff5eb', '#FFFFFF', '#f3f7fd', '#e8effa', '#dce8f8', '#d0e0f6', '#c4d8f3'],
    ClassicRedWhiteGreenLight11: ['#ffb2b6', '#ffc2c5', '#ffd1d3', '#ffe0e1', '#fff0f0', '#FFFFFF', '#f1faed', '#e3f5db', '#d5f0ca', '#c6ebb8', '#b7e6a7'],
    ClassicRedGreenLight11: ['#ffb2b6', '#fcbdc0', '#f8c7c9', '#f2d1d2', '#ecdbdc', '#e5e5e5', '#dde6d9', '#d4e6cc', '#cae6c0', '#c1e6b4', '#b7e6a7'],
};

var colorschemes = {
    brewer: brewer,
    office: office,
    tableau: tableau,
};

const { get } = lodash;
const getColorschemeColors = (colorscheme, customColorSchemes) => {
    var _a;
    if (!colorscheme) {
        return ['#37A2DA', '#32C5E9', '#67E0E3', '#9FE6B8', '#FFDB5C', '#ff9f7f', '#fb7293', '#E062AE', '#E690D1', '#e7bcf3', '#9d96f5', '#8378EA', '#96BFFF'];
    }
    if (colorscheme.includes('custom') && customColorSchemes) {
        return ((_a = customColorSchemes.find(({ id }) => id === colorscheme)) === null || _a === void 0 ? void 0 : _a.colors) || [];
    }
    return get(colorschemes, colorscheme);
};

/**
 * Locale-related helper functions
 */
/**
 * Get the first day of the week for a given locale.
 * Returns 0 for Sunday, 1 for Monday, etc.
 *
 * @param locale - BCP 47 language tag (e.g., 'en-US', 'hu-HU')
 * @returns The first day of the week (0-6)
 */
function getWeekStartDay(locale) {
    const l = new Intl.Locale(locale);
    // @ts-ignore - getWeekInfo is a newer API
    if (l.getWeekInfo) {
        // @ts-ignore
        return (l.getWeekInfo().firstDay % 7);
    }
    // @ts-ignore - weekInfo is deprecated but still supported in some browsers
    if (l.weekInfo && l.weekInfo.firstDay !== undefined) {
        // @ts-ignore
        return (l.weekInfo.firstDay % 7);
    }
    return 0;
}

var index$7 = /*#__PURE__*/Object.freeze({
  __proto__: null,
  Attachment: Attachment,
  colorschemes: colorschemes,
  getColorschemeColors: getColorschemeColors,
  getWeekStartDay: getWeekStartDay
});

class BasicChartOptions extends ChartOptions {
    constructor(o) {
        super(o);
        this.labelColumn = '';
        this.dataColumns = [];
        if (!o)
            return;
        Apply(this, o, String, 'labelColumn');
        if (o.dataColumns) {
            this.dataColumns = o.dataColumns || [];
        }
    }
    getChartConfiguration(dataframes, meta) {
        const { themeVariables = {} } = meta;
        const { labels, datasets = [] } = this.getData(dataframes[0], dataframes);
        const options = {
            series: [],
            xAxis: [],
            yAxis: [],
            tooltip: {
                show: true,
                appendToBody: true,
            },
        };
        if (['pie', 'doughnut'].includes(this.type)) {
            const startRadius = this.type === 'doughnut' ? 40 : 0;
            const endRadius = 80;
            const radiusLength = (endRadius - startRadius) / (datasets.length || 1);
            options.tooltip.trigger = 'item';
            options.series = datasets.map(({ label, data }, index) => {
                const sr = startRadius + (index * radiusLength);
                const er = startRadius + ((index + 1) * radiusLength);
                return {
                    name: label,
                    type: 'pie',
                    radius: [`${sr}%`, `${er}%`],
                    center: ['50%', '55%'],
                    tooltip: {
                        formatter: (params) => {
                            return `${params.seriesName}<br>${params.marker}${params.name}<span style="float: right; margin-left: 20px">${params.value} (${params.percent}%)</span>`;
                        },
                        appendToBody: true,
                    },
                    label: {
                        show: this.tooltips.showAlways,
                        position: 'inside',
                        align: 'center',
                        verticalAlign: 'middle',
                        formatter: '{c} ({d}%)',
                    },
                    itemStyle: {
                        borderRadius: 5,
                        borderColor: themeVariables.white,
                        borderWidth: 1,
                    },
                    emphasis: {
                        itemStyle: {
                            shadowBlur: 10,
                            shadowOffsetX: 0,
                            shadowColor: 'rgba(0, 0, 0, 0.5)',
                        },
                    },
                    data: labels.map((name, i) => {
                        return { name, value: data[i] };
                    }),
                    top: this.offset.default ? undefined : this.offset.top,
                    right: this.offset.default ? undefined : this.offset.right,
                    bottom: this.offset.default ? undefined : this.offset.bottom,
                    left: this.offset.default ? undefined : this.offset.left,
                };
            });
        }
        else if (['bar', 'line'].includes(this.type)) {
            options.tooltip.trigger = 'axis';
            const { label: xLabel, type: xType = 'category', labelRotation: xLabelRotation = 0, } = this.xAxis;
            options.xAxis = [
                {
                    name: xLabel,
                    nameLocation: 'center',
                    nameGap: 30,
                    type: xType,
                    data: labels,
                    axisLabel: {
                        interval: 0,
                        overflow: 'truncate',
                        hideOverlap: true,
                        rotate: xLabelRotation,
                    },
                    axisTick: {
                        show: false,
                    },
                    axisLine: {
                        show: false,
                    },
                },
            ];
            options.grid = {
                top: this.offset.default ? (this.title ? 70 : 45) : this.offset.top,
                right: this.offset.default ? 30 : this.offset.right,
                bottom: this.offset.default ? (xLabel ? 30 : 25) : this.offset.bottom,
                left: this.offset.default ? 30 : this.offset.left,
                containLabel: true,
            };
            const { label: yLabel, labelRotation: yLabelRotation = 0, type: yType = 'linear', position = 'left', labelPosition = 'end', beginAtZero, min, max, } = this.yAxis;
            const tempYAxis = {
                name: yLabel,
                type: yType === 'linear' ? 'value' : 'log',
                position,
                nameGap: labelPosition === 'center' ? 25 : 7,
                nameLocation: labelPosition,
                min: beginAtZero ? 0 : min || undefined,
                max: max || undefined,
                axisLabel: {
                    interval: 0,
                    overflow: 'truncate',
                    hideOverlap: true,
                    rotate: yLabelRotation,
                },
                axisLine: {
                    show: false,
                    onZero: false,
                },
                splitLine: {
                    lineStyle: {
                        color: [themeVariables['extra-light']],
                    },
                },
                nameTextStyle: {
                    align: labelPosition === 'center' ? 'center' : position,
                    padding: labelPosition !== 'center' ? (position === 'left' ? [0, 0, 2, -20] : [0, -20, 2, 0]) : undefined,
                },
            };
            // If we provide undefined, log scale breaks
            if (tempYAxis.type === 'log') {
                delete tempYAxis.min;
                delete tempYAxis.max;
            }
            options.yAxis = [tempYAxis];
            options.series = datasets.map(({ label, data, stack }) => {
                return {
                    name: label,
                    type: this.type,
                    smooth: true,
                    areaStyle: {},
                    left: 'left',
                    stack,
                    label: {
                        show: this.tooltips.showAlways,
                        position: 'inside',
                        align: 'center',
                        verticalAlign: 'middle',
                    },
                    data: xType === 'time' ? labels.map((name, i) => {
                        return [moment(name).valueOf() || undefined, data[i]];
                    }) : data,
                };
            });
        }
        return Object.assign({ animation: !this.noAnimation, title: {
                text: this.title,
                left: 'center',
                textStyle: {
                    fontFamily: themeVariables['font-regular'],
                    color: themeVariables.black,
                    fontSize: 16,
                },
            }, color: getColorschemeColors(this.colorScheme), textStyle: {
                fontFamily: themeVariables['font-regular'],
                overflow: 'break',
                color: themeVariables.black,
            }, legend: {
                show: !this.legend.hide,
                type: this.legend.scrollable ? 'scroll' : 'plain',
                top: (this.legend.position.default ? (this.title ? 25 : undefined) : this.legend.position.top) || undefined,
                right: (this.legend.position.default ? undefined : this.legend.position.right) || undefined,
                bottom: (this.legend.position.default ? undefined : this.legend.position.bottom) || undefined,
                left: (this.legend.position.default ? this.legend.align || 'center' : this.legend.position.left) || 'auto',
                orient: this.legend.orientation || 'horizontal',
                textStyle: {
                    color: themeVariables.black,
                },
                pageTextStyle: {
                    color: themeVariables.black,
                },
                pageIconColor: themeVariables.black,
                pageIconInactiveColor: themeVariables.light,
            } }, options);
    }
    getColIndex(dataframe, col) {
        if (!dataframe || !dataframe.columns)
            return -1;
        return dataframe.columns.findIndex(({ name }) => name === col);
    }
    getData(localDataframe, dataframes) {
        const datasets = [];
        let labels = [];
        if (localDataframe && dataframes) {
            // Get datasets
            if (this.dataColumns.length && localDataframe.rows) {
                for (const { name, stack, label } of this.dataColumns) {
                    // Assume localDataframe has the dataColumn
                    let columnIndex = this.getColIndex(localDataframe, name);
                    // If dataColumn is in localDataframe, then set that value
                    const data = localDataframe.rows.map(r => {
                        return columnIndex < 0 ? undefined : r[columnIndex];
                    });
                    if (columnIndex < 0) {
                        dataframes.slice(1).forEach(df => {
                            const { relColumn = '', refValue = '' } = df;
                            // Get column that is referenced by relColumn
                            const relColumnIndex = this.getColIndex(localDataframe, relColumn);
                            if (relColumnIndex < 0) {
                                throw new Error(`Column ${relColumn} not found`);
                            }
                            if (!localDataframe.rows) {
                                throw new Error('Local rows not found');
                            }
                            // Get row index that matches refValue
                            const refRowIndex = localDataframe.rows.findIndex(row => row[relColumnIndex] === refValue);
                            if (refRowIndex < 0) {
                                throw new Error(`Row that matches refRowIndex ${refValue} not found`);
                            }
                            columnIndex = this.getColIndex(df, name);
                            if (columnIndex < 0) {
                                throw new Error(`Column ${name} not found`);
                            }
                            else if (df.rows) {
                                data[refRowIndex] = df.rows[0][columnIndex];
                            }
                        });
                    }
                    datasets.push({
                        label: label || name,
                        data,
                        stack,
                    });
                }
            }
            // Get labels, if dimensions type is not time
            if (this.labelColumn && localDataframe) {
                const columnIndex = this.getColIndex(localDataframe, this.labelColumn);
                if (columnIndex < 0) {
                    throw new Error(`Column ${this.labelColumn} not found`);
                }
                if (localDataframe.rows) {
                    for (const row of localDataframe.rows) {
                        const label = row[columnIndex] || (!this.xAxis.skipMissing ? this.xAxis.defaultValue : undefined);
                        labels.push(label);
                    }
                }
            }
            if (this.xAxis.skipMissing) {
                labels.forEach((label, index) => {
                    if (!label) {
                        datasets.forEach(ds => {
                            ds.data.splice(index, 1);
                        });
                    }
                });
                labels = labels.filter(label => label);
            }
        }
        return { datasets, labels };
    }
}
ChartOptionsRegistry.set('bar', BasicChartOptions);
ChartOptionsRegistry.set('line', BasicChartOptions);
ChartOptionsRegistry.set('pie', BasicChartOptions);
ChartOptionsRegistry.set('doughnut', BasicChartOptions);

class FunnelChartOptions extends ChartOptions {
    constructor(o) {
        super(o);
        this.labelColumn = '';
        this.dataColumns = [];
        if (!o)
            return;
        Apply(this, o, String, 'labelColumn');
        if (o.dataColumns) {
            this.dataColumns = o.dataColumns || [];
        }
    }
    getChartConfiguration(dataframes, meta) {
        const { themeVariables = {} } = meta;
        const labels = this.getLabels(dataframes[0]);
        const { data = [] } = this.getDatasets(dataframes[0], dataframes) || {};
        const colors = getColorschemeColors(this.colorScheme);
        return {
            animation: !this.noAnimation,
            title: {
                text: this.title,
                left: 'center',
                textStyle: {
                    fontFamily: themeVariables['font-regular'],
                    color: themeVariables.black,
                    fontSize: 16,
                },
            },
            textStyle: {
                fontFamily: themeVariables['font-regular'],
            },
            tooltip: {
                show: true,
                trigger: 'item',
                formatter: (params) => {
                    return `${params.seriesName}<br>${params.marker}${params.name}<span style="float: right; margin-left: 20px">${params.value} (${params.percent}%)</span>`;
                },
                appendToBody: true,
            },
            legend: {
                show: !this.legend.hide,
                type: this.legend.scrollable ? 'scroll' : 'plain',
                top: (this.legend.position.default ? (this.title ? 25 : undefined) : this.legend.position.top) || undefined,
                right: (this.legend.position.default ? undefined : this.legend.position.right) || undefined,
                bottom: (this.legend.position.default ? undefined : this.legend.position.bottom) || undefined,
                left: (this.legend.position.default ? this.legend.align || 'center' : this.legend.position.left) || 'auto',
                orient: this.legend.orientation || 'horizontal',
                textStyle: {
                    color: themeVariables.black,
                },
                pageTextStyle: {
                    color: themeVariables.black,
                },
                pageIconColor: themeVariables.black,
                pageIconInactiveColor: themeVariables.light,
            },
            series: [
                {
                    type: 'funnel',
                    name: this.labelColumn,
                    sort: 'descending',
                    width: '90%',
                    label: {
                        show: this.tooltips.showAlways,
                        position: 'inside',
                        align: 'center',
                        verticalAlign: 'middle',
                        formatter: '{c} ({d}%)',
                    },
                    data: labels.map((name, i) => {
                        return { name, value: data[i], itemStyle: { color: colors[i] } };
                    }),
                    top: this.offset.default ? (this.title ? 60 : 35) : this.offset.top,
                    right: this.offset.default ? '5%' : this.offset.right,
                    bottom: this.offset.default ? '5%' : this.offset.bottom,
                    left: this.offset.default ? '5%' : this.offset.left,
                },
            ],
        };
    }
    getColIndex(dataframe, col) {
        if (!dataframe || !dataframe.columns)
            return -1;
        return dataframe.columns.findIndex(({ name }) => name === col);
    }
    getLabels(localDataframe) {
        const labels = [];
        if (this.labelColumn && localDataframe) {
            const columnIndex = this.getColIndex(localDataframe, this.labelColumn);
            if (columnIndex < 0) {
                throw new Error(`Column ${this.labelColumn} not found`);
            }
            if (localDataframe.rows) {
                for (const row of localDataframe.rows) {
                    labels.push(row[columnIndex]);
                }
            }
        }
        return labels;
    }
    getDatasets(localDataframe, dataframes) {
        const chartDataset = [];
        if (localDataframe && dataframes) {
            if (this.dataColumns.length && localDataframe.rows) {
                // Create dataset for each dataColumn
                for (const { name } of this.dataColumns) {
                    // Assume localDataframe has the dataColumn
                    let columnIndex = this.getColIndex(localDataframe, name);
                    // If dataColumn is in localDataframe, then set that value
                    const data = localDataframe.rows.map(r => {
                        return columnIndex < 0 ? undefined : parseFloat(r[columnIndex] || '0') || 0;
                    });
                    // Otherwise check other dataframes for that columnn
                    if (columnIndex < 0) {
                        dataframes.slice(1).forEach(df => {
                            const { relColumn = '', refValue = '' } = df;
                            // Get column that is referenced by relColumn
                            const relColumnIndex = this.getColIndex(localDataframe, relColumn);
                            if (relColumnIndex < 0) {
                                throw new Error(`Column ${relColumn} not found`);
                            }
                            if (!localDataframe.rows) {
                                throw new Error('Local rows not found');
                            }
                            // Get row index that matches refValue
                            const refRowIndex = localDataframe.rows.findIndex(row => row[relColumnIndex] === refValue);
                            if (refRowIndex < 0) {
                                throw new Error(`Row that matches refRowIndex ${refValue} not found`);
                            }
                            columnIndex = this.getColIndex(df, name);
                            if (columnIndex < 0) {
                                throw new Error(`Column ${name} not found`);
                            }
                            else if (df.rows) {
                                data[refRowIndex] = parseFloat(df.rows[0][columnIndex] || '0') || 0;
                            }
                        });
                    }
                    chartDataset.push({
                        label: name,
                        data,
                    });
                }
            }
        }
        return chartDataset[0];
    }
}
ChartOptionsRegistry.set('funnel', FunnelChartOptions);

const kind$v = 'Table';
const defaults$v = Object.freeze({
    source: '',
    datasources: [],
    columns: {},
    striped: false,
    bordered: false,
    borderless: false,
    small: false,
    hover: false,
    dark: false,
    fixed: false,
    responsive: true,
    noCollapse: false,
    headVariant: null,
    tableVariant: '',
});
class DisplayElementTable extends DisplayElement {
    constructor(i) {
        super(i);
        this.kind = kind$v;
        this.options = Object.assign({}, defaults$v);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, String, 'headVariant', 'tableVariant', 'source');
        Apply(this.options, o, Boolean, 'striped', 'bordered', 'borderless', 'small', 'hover', 'dark', 'fixed', 'responsive', 'noCollapse');
        if (o.datasources) {
            this.options.datasources = o.datasources || [];
        }
        if (o.columns) {
            this.options.columns = o.columns || [];
        }
    }
    reportDefinitions(definition = {}) {
        if (typeof this.options.source === 'object') {
            // @todo allow implicit sources
            throw new Error('table source must be provided as a reference');
        }
        const dataframes = [];
        this.options.datasources.forEach(({ name = '', filter, sort, paging }) => {
            var _a, _b;
            const df = {
                name: this.elementID,
                source: this.options.source,
                ref: name,
                filter,
                sort,
                paging,
            };
            const relatedDefinition = definition[name];
            if (relatedDefinition) {
                df.sort = (relatedDefinition.sort ? relatedDefinition.sort : sort) || undefined;
                if (relatedDefinition.filter && ((_a = relatedDefinition.filter) === null || _a === void 0 ? void 0 : _a.ref)) {
                    // If element and scenario have filter AND them together
                    if (filter && filter.ref) {
                        df.filter = {
                            ref: 'and',
                            args: [
                                filter,
                                relatedDefinition.filter,
                            ],
                        };
                    }
                    else {
                        df.filter = relatedDefinition.filter;
                    }
                }
                if (relatedDefinition.paging || paging) {
                    df.paging = Object.assign(Object.assign({}, (paging || {})), (relatedDefinition.paging || {}));
                }
            }
            if ((_b = df.paging) === null || _b === void 0 ? void 0 : _b.limit) {
                df.paging.limit = parseInt(df.paging.limit);
            }
            dataframes.push(df);
        });
        return { dataframes };
    }
}
Registry$2.set(kind$v, DisplayElementTable);

const kind$u = 'Text';
const defaults$u = Object.freeze({
    value: 'Sample text...',
});
class DisplayElementText extends DisplayElement {
    constructor(i) {
        super(i);
        this.kind = kind$u;
        this.options = Object.assign({}, defaults$u);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, String, 'value');
    }
}
Registry$2.set(kind$u, DisplayElementText);

const kind$t = 'Metric';
const defaults$t = Object.freeze({
    source: '',
    datasources: [],
    valueColumn: '',
    format: '',
    prefix: '',
    suffix: '',
    color: '#0B344E',
    backgroundColor: '#FFFFFF00',
});
class DisplayElementMetric extends DisplayElement {
    constructor(i) {
        super(i);
        this.kind = kind$t;
        this.options = Object.assign({}, defaults$t);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, String, 'source', 'valueColumn', 'format', 'prefix', 'suffix', 'color', 'backgroundColor');
        if (o.datasources) {
            this.options.datasources = o.datasources || [];
        }
    }
    reportDefinitions(definition = {}) {
        if (typeof this.options.source === 'object') {
            // @todo allow implicit sources
            throw new Error('metric source must be provided as a reference');
        }
        const dataframes = [];
        this.options.datasources.forEach(({ name = '', filter, sort }) => {
            var _a;
            const df = {
                name: this.elementID,
                source: this.options.source,
                ref: name,
                filter,
                sort,
            };
            const relatedDefinition = definition[name];
            if (relatedDefinition) {
                df.sort = (relatedDefinition.sort ? relatedDefinition.sort : sort) || undefined;
                if (relatedDefinition.filter && ((_a = relatedDefinition.filter) === null || _a === void 0 ? void 0 : _a.ref)) {
                    // If element and scenario have filter AND them together
                    if (filter && filter.ref) {
                        df.filter = {
                            ref: 'and',
                            args: [
                                filter,
                                relatedDefinition.filter,
                            ],
                        };
                    }
                    else {
                        df.filter = relatedDefinition.filter;
                    }
                }
            }
            dataframes.push(df);
        });
        return { dataframes };
    }
}
Registry$2.set(kind$t, DisplayElementMetric);

function DisplayElementMaker(i) {
    const DisplayElementTemp = Registry$2.get(i.kind);
    if (DisplayElementTemp === undefined) {
        throw new Error(`unknown display element kind '${i.kind}'`);
    }
    if (i instanceof DisplayElement) {
        // Get rid of the references
        i = JSON.parse(JSON.stringify(i));
    }
    return new DisplayElementTemp(i);
}

const defaultXYWH$1 = () => [0, 0, 20, 15];
class Block {
    constructor(p) {
        this.blockID = NoID;
        this.title = '';
        this.description = '';
        this.layout = 'horizontal';
        this.elements = [];
        this.xywh = defaultXYWH$1();
        if (!p)
            return;
        Apply(this, p, String, 'title', 'description', 'layout');
        Apply(this, p, CortezaID, 'blockID');
        if (p.xywh) {
            if (!Array.isArray(p.xywh)) {
                throw new Error('xywh must be an array');
            }
            if (p.xywh.length !== 4) {
                throw new Error('xywh must have 4 elements');
            }
            this.xywh = p.xywh;
        }
        if (p.elements) {
            this.elements = [];
            if (AreObjectsOf(p.elements, 'kind')) {
                this.elements = p.elements.map((e) => DisplayElementMaker(e));
            }
        }
    }
}

function StepFactory(step) {
    const k = Object.keys(step)[0];
    switch (k) {
        case 'load':
            return step;
        case 'link':
            return step;
        case 'join':
            return step;
        case 'aggregate':
            return step;
        default:
            throw new Error('unknown step: ' + k);
    }
}

var index$6 = /*#__PURE__*/Object.freeze({
  __proto__: null,
  Block: Block,
  ChartOptionsMaker: ChartOptionsMaker,
  DisplayElement: DisplayElement,
  DisplayElementChart: DisplayElementChart,
  DisplayElementMaker: DisplayElementMaker,
  DisplayElementMetric: DisplayElementMetric,
  DisplayElementRegistry: Registry$2,
  DisplayElementTable: DisplayElementTable,
  DisplayElementText: DisplayElementText,
  StepFactory: StepFactory
});

class Report {
    constructor(r) {
        this.reportID = NoID;
        this.handle = '';
        this.meta = {
            name: '',
            description: '',
        };
        this.sources = [];
        this.blocks = [];
        this.scenarios = [];
        this.labels = {};
        this.createdAt = undefined;
        this.createdBy = undefined;
        this.updatedAt = undefined;
        this.updatedBy = undefined;
        this.deletedAt = undefined;
        this.deletedBy = undefined;
        this.canReadReport = false;
        this.canUpdateReport = false;
        this.canDeleteReport = false;
        this.canGrant = false;
        this.canRunReport = false;
        this.apply(r);
    }
    apply(r) {
        Apply(this, r, CortezaID, 'reportID');
        Apply(this, r, String, 'handle');
        if (r && IsOf(r, 'meta')) {
            this.meta = r.meta;
        }
        this.sources = [];
        for (const s of (r === null || r === void 0 ? void 0 : r.sources) || []) {
            s.step = s.step;
            this.sources.push(s);
        }
        if (r === null || r === void 0 ? void 0 : r.blocks) {
            this.blocks = [];
            for (const p of r.blocks) {
                this.blocks.push(new Block(p));
            }
        }
        if (r === null || r === void 0 ? void 0 : r.scenarios) {
            this.scenarios = [];
            for (const s of r.scenarios) {
                this.scenarios.push(s);
            }
        }
        if (IsOf(r, 'labels')) {
            this.labels = Object.assign({}, r.labels);
        }
        Apply(this, r, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, r, CortezaID, 'createdBy', 'updatedBy', 'deletedBy');
        Apply(this, r, Boolean, 'canReadReport', 'canUpdateReport', 'canDeleteReport', 'canGrant', 'canRunReport');
    }
    clone() {
        return new Report(JSON.parse(JSON.stringify(this)));
    }
}

const { merge: merge$a } = lodash;
class DalConnection {
    constructor(dc) {
        this.connectionID = NoID;
        this.handle = '';
        this.type = 'corteza::system:dal-connection';
        this.meta = {
            name: '',
            ownership: '',
            location: {
                properties: { name: '' },
                geometry: {
                    coordinates: [],
                    type: '',
                },
            },
            properties: {
                dataAtRestEncryption: {
                    enabled: false,
                    notes: '',
                },
                dataAtRestProtection: {
                    enabled: false,
                    notes: '',
                },
                dataAtTransitEncryption: {
                    enabled: false,
                    notes: '',
                },
                dataRestoration: {
                    enabled: false,
                    notes: '',
                },
            },
        };
        this.config = {
            privacy: { sensitivityLevelID: NoID },
            dal: {},
        };
        this.issues = [];
        this.labels = [];
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.createdBy = NoID;
        this.updatedBy = NoID;
        this.deletedBy = NoID;
        this.canDeleteConnection = false;
        this.canManageDalConfig = false;
        this.apply(dc);
    }
    apply(dc) {
        Apply(this, dc, CortezaID, 'connectionID');
        Apply(this, dc, String, 'handle', 'type');
        Apply(this, dc, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, dc, CortezaID, 'createdBy', 'updatedBy', 'deletedBy');
        Apply(this, dc, Boolean, 'canDeleteConnection', 'canManageDalConfig');
        if (IsOf(dc, 'meta')) {
            this.meta = merge$a(this.meta, dc.meta);
        }
        if (IsOf(dc, 'config')) {
            this.config = Object.assign({}, dc.config);
            if (this.connectionID !== NoID && this.canManageDalConfig) {
                this.config = Object.assign({ dal: {
                        type: 'corteza::dal:connection:dsn',
                        params: { dsn: '' },
                        modelIdent: '',
                        modelIdentCheck: [],
                    } }, dc.config);
                if (!this.config.privacy.sensitivityLevelID) {
                    this.config.privacy = {
                        sensitivityLevelID: NoID,
                    };
                }
            }
        }
        if (dc === null || dc === void 0 ? void 0 : dc.issues) {
            this.issues = [];
            for (const i of dc.issues) {
                this.issues.push(i);
            }
        }
        if (dc === null || dc === void 0 ? void 0 : dc.labels) {
            this.labels = [];
            for (const l of dc.labels) {
                this.labels.push(l);
            }
        }
    }
    clone() {
        return new DalConnection(JSON.parse(JSON.stringify(this)));
    }
}

class AuthClient {
    constructor(o) {
        this.authClientID = NoID;
        this.handle = '';
        this.scope = 'profile api';
        this.redirectURI = '';
        this.validGrant = 'authorization_code';
        this.meta = {
            name: '',
            description: '',
        };
        this.security = {
            userGroup: NoID,
            impersonateUser: NoID,
            permittedRoles: [],
            prohibitedRoles: [],
            forcedRoles: [],
        };
        this.enabled = true;
        this.trusted = false;
        this.validFrom = undefined;
        this.expiresAt = undefined;
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.createdBy = NoID;
        this.updatedBy = NoID;
        this.deletedBy = NoID;
        this.canDeleteAuthClient = false;
        this.canGrant = false;
        this.canUpdateAuthClient = false;
        this.apply(o);
    }
    apply(o) {
        Apply(this, o, CortezaID, 'authClientID');
        Apply(this, o, ISO8601Date, 'validFrom', 'expiresAt', 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, o, String, 'handle', 'scope', 'redirectURI', 'validGrant');
        Apply(this, o, Boolean, 'enabled', 'trusted', 'canDeleteAuthClient', 'canGrant', 'canUpdateAuthClient');
        if (IsOf(o, 'meta')) {
            this.meta = Object.assign({}, o.meta);
        }
        if (IsOf(o, 'security')) {
            this.security = Object.assign(Object.assign({}, this.security), o.security);
        }
        Apply(this, o, CortezaID, 'createdBy', 'updatedBy', 'deletedBy');
    }
    clone() {
        return new AuthClient(JSON.parse(JSON.stringify(this)));
    }
}

var NotificationKind;
(function (NotificationKind) {
    NotificationKind["Simple"] = "simple";
    NotificationKind["Record"] = "record";
})(NotificationKind || (NotificationKind = {}));
class Notification {
    constructor(n) {
        this.notificationID = NoID;
        this.kind = NotificationKind.Simple;
        this.config = { simple: { title: '', description: '' } };
        this.recipient = NoID;
        this.createdBy = NoID;
        this.createdAt = undefined;
        this.readAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.apply(n);
    }
    apply(n) {
        if (!n)
            return;
        Apply(this, n, CortezaID, 'notificationID', 'recipient', 'createdBy');
        Apply(this, n, String, 'kind');
        Apply(this, n, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt', 'readAt');
        if (IsOf(n, 'config')) {
            this.config = n.config[this.kind] || {};
        }
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.notificationID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'system:notification';
    }
    clone() {
        return new Notification(JSON.parse(JSON.stringify(this)));
    }
}

class Revision {
    constructor(r) {
        this.changeID = '';
        this.timestamp = undefined;
        this.resource = '';
        this.revision = 0;
        this.operation = 'created';
        this.status = '';
        this.userID = NoID;
        this.changes = [];
        this.comment = '';
        this.deletedAt = undefined;
        this.deletedBy = NoID;
        this.record = undefined;
        this.apply(r);
    }
    apply(r) {
        if (!r)
            return;
        Apply(this, r, String, 'changeID', 'resource');
        Apply(this, r, CortezaID, 'userID', 'deletedBy');
        Apply(this, r, ISO8601Date, 'timestamp', 'deletedAt');
        Apply(this, r, Number, 'revision');
        Apply(this, r, String, 'operation', 'status', 'comment');
        Apply(this, r, (v) => v, 'record');
        if (IsOf(r, 'changes') && Array.isArray(r.changes)) {
            this.changes = r.changes.map(c => ({
                key: c.key || '',
                old: Array.isArray(c.old) ? c.old : [],
                new: Array.isArray(c.new) ? c.new : [],
            }));
        }
    }
    get resourceIdentifier() {
        return this.resource || `system:revision:${this.changeID}`;
    }
    get isDraft() {
        return this.status === 'draft';
    }
    clone() {
        return new Revision(JSON.parse(JSON.stringify(this)));
    }
}

// @todo refactor this into more compose-like event structure (see compose/events.ts
function SystemEvent(eventType = onManual) {
    return GenericEventMaker({ resourceType: 'system' }, eventType, () => true, {});
}
// @todo refactor this into more compose-like event structure (see compose/events.ts
function UserEvent(user, eventType = onManual) {
    return GenericEventMaker(user, eventType, function (c) {
        switch (c.Name()) {
            case 'user':
            case 'user.handle':
                return c.Match(user.handle);
            case 'user.email':
                return c.Match(user.email);
        }
        return false;
    }, { user });
}
// @todo refactor this into more compose-like event structure (see compose/events.ts
function RoleEvent(role, eventType = onManual) {
    return GenericEventMaker(role, eventType, function (c) {
        switch (c.Name()) {
            case 'role':
            case 'role.handle':
                return c.Match(role.handle);
            case 'role.name':
                return c.Match(role.name);
        }
        return false;
    }, { role });
}
/**
 * Returns handler that routes onManual events for server script to the system API
 *
 * See makeAutomationScriptsRegistrator
 *
 * @param api
 * @return function
 */
function TriggerSystemServerScriptOnManual(api) {
    return (ev, script) => {
        var _a, _b;
        const params = { script, args: ev.args };
        const { userID } = (_a = ev.args) === null || _a === void 0 ? void 0 : _a.user;
        const { roleID } = (_b = ev.args) === null || _b === void 0 ? void 0 : _b.role;
        switch (ev.resourceType) {
            case 'system':
                return api.automationTriggerScript(Object.assign({}, params));
            case 'system:user':
                return api.userTriggerScript(Object.assign({ userID }, params)).then(rval => new User(rval));
            case 'system:role':
                return api.roleTriggerScript(Object.assign({ roleID }, params)).then(rval => new Role(rval));
            default:
                throw Error(`cannot trigger server script: unknown resource type '${ev.resourceType}'`);
        }
    };
}

var index$5 = /*#__PURE__*/Object.freeze({
  __proto__: null,
  Application: Application,
  AuthClient: AuthClient,
  DalConnection: DalConnection,
  Notification: Notification,
  Reminder: Reminder,
  Report: Report,
  Revision: Revision,
  Role: Role,
  RoleEvent: RoleEvent,
  SinkRequest: SinkRequest,
  SinkResponse: SinkResponse,
  SystemEvent: SystemEvent,
  Template: Template,
  TriggerSystemServerScriptOnManual: TriggerSystemServerScriptOnManual,
  User: User,
  UserEvent: UserEvent,
  UserGroup: UserGroup
});

/**
 * Helpers to determine if specific object looks like the type we are interested in.
 * It does not rely on instanceof, because of bundling issues.
 */
function isUser(o) {
    return o && !!o.userID;
}
function isRole(o) {
    return o && !!o.roleID;
}
/**
 * SystemHelper provides layer over System API and utilities that simplify automation script writing
 */
class SystemHelper {
    constructor(ctx) {
        this.SystemAPI = ctx.SystemAPI;
        this.$user = ctx.$user;
        this.$role = ctx.$role;
        this.$application = ctx.$application;
    }
    /**
     * Searches for users
     *
     * @example
     * System.findUsers('some-joe').then(({ set }) => {
     *   // do something with users (User[]) in set
     * })
     *
     * @param filter - filter object (or filtering conditions when string)
     * @property filter.query - Find %query% in email, handle, username, name...
     * @property filter.username - Filter by username
     * @property filter.handle - Filter by handle
     * @property filter.email - Filter by email
     * @property filter.kind - Filter by kind ('normal' - default, 'bot')
     * @property filter.incDeleted - Include deleted users
     * @property filter.incSuspended - Include suspended users
     * @property filter.sort - Sort results
     * @property filter.perPage - max returned records per page
     * @property filter.page - page to return (1-based)
     */
    findUsers(filter) {
        return __awaiter(this, void 0, void 0, function* () {
            if (typeof filter === 'string') {
                filter = { query: filter };
            }
            return this.SystemAPI
                .userList(filter || {})
                .then(res => {
                res.set = res.set.map(u => new User(u));
                return res;
            });
        });
    }
    /**
     * Finds user by ID
     *
     * @example
     * System.findUserByID()
     *
     * @param user
     */
    findUserByID(user) {
        return __awaiter(this, void 0, void 0, function* () {
            const userID = extractID(user, 'userID');
            return this.SystemAPI.userRead({ userID }).then(u => new User(u));
        });
    }
    /**
     * Finds user by email
     *
     * @example
     * System.findUserByEmail('name@example.tld').then(user => {
     *   // do something with user
     * })
     *
     * @param email
     */
    findUserByEmail(email) {
        return __awaiter(this, void 0, void 0, function* () {
            return this.findUsers({ email }).then(res => {
                if (!Array.isArray(res.set) || res.set.length === 0) {
                    throw new Error('user not found');
                }
                return new User(res.set[0]);
            });
        });
    }
    /**
     * Finds user by handle
     *
     * @example
     * System.findUserByHandle('some-handle').then(user => {
     *   // do something with user
     * })
     *
     * @param handle
     */
    findUserByHandle(handle) {
        return __awaiter(this, void 0, void 0, function* () {
            return this.findUsers({ handle }).then(res => {
                if (!Array.isArray(res.set) || res.set.length === 0 || !res.set) {
                    throw new Error('user not found');
                }
                return new User(res.set[0]);
            });
        });
    }
    /**
     * Updates or creates user
     *
     * @example
     * System.findUserByHandle('some-handle').then(user => {
     *   user.handle = 'better-handle'
     *   return System.saveUser(user)
     * })
     *
     * @param user
     */
    saveUser(user) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(user).then(user => {
                if (isFresh(user.userID)) {
                    return this.SystemAPI.userCreate(kv(user)).then(user => new User(user));
                }
                else {
                    return this.SystemAPI.userUpdate(kv(user)).then(user => new User(user));
                }
            });
        });
    }
    /**
     * Sets/updates password for the user
     *
     * @example
     * System.findUserByHandle('some-handle').then(user => {
     *   user.handle = 'better-handle'
     *   return System.saveUser(user)
     * })
     *
     * @param password
     * @param user
     */
    setPassword(password_1) {
        return __awaiter(this, arguments, void 0, function* (password, user = this.$user) {
            return this.resolveUser(user).then(user => {
                const { userID } = user;
                if (isFresh(userID)) {
                    throw new Error('Cannot set password for non existing user');
                }
                return this.SystemAPI.userSetPassword({ password, userID }).then(u => new User(u));
            });
        });
    }
    /**
     * Deletes user
     *
     * @example
     * System.findUserByHandle('soon-to-be-deleted').then(user => {
     *   return System.deleteUser(user)
     * })
     *
     * @param user
     */
    deleteUser(user) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(user).then(user => {
                const userID = extractID(user, 'userID');
                if (!isFresh(userID)) {
                    return this.SystemAPI.userDelete({ userID });
                }
            });
        });
    }
    /**
     * Searches for roles
     *
     * @param filter
     */
    findRoles(filter) {
        return __awaiter(this, void 0, void 0, function* () {
            if (typeof filter === 'string') {
                filter = { query: filter };
            }
            return this.SystemAPI
                .roleList(filter || {})
                .then(res => {
                res.set = res.set.map(r => new Role(r));
                return res;
            });
        });
    }
    /**
     * Finds user by ID
     *
     * @param role
     */
    findRoleByID(role) {
        return __awaiter(this, void 0, void 0, function* () {
            const roleID = extractID(role, 'roleID');
            return this.SystemAPI.roleRead({ roleID }).then(r => new Role(r));
        });
    }
    /**
     * Finds role by handle
     *
     * @example
     * System.findRoleByHandle('some-handle').then(user => {
     *   // do something with role
     * })
     *
     * @param handle
     */
    findRoleByHandle(handle) {
        return __awaiter(this, void 0, void 0, function* () {
            return this.findRoles(handle).then(res => {
                if (!Array.isArray(res.set) || res.set.length === 0 || !res.set) {
                    throw new Error('role not found');
                }
                return new Role(res.set[0]);
            });
        });
    }
    /**
     *
     * @param role
     */
    saveRole(role) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(role).then(role => {
                if (isFresh(role.roleID)) {
                    return this.SystemAPI.roleCreate(kv(role)).then(role => new Role(role));
                }
                else {
                    return this.SystemAPI.roleUpdate(kv(role)).then(role => new Role(role));
                }
            });
        });
    }
    /**
     * Deletes a role
     *
     * @example
     * System.findUserByHandle('soon-to-be-deleted').then(user => {
     *   return System.deleteUser(user)
     * })
     *
     * @param role
     */
    deleteRole(role) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(role).then(role => {
                const roleID = extractID(role, 'roleID');
                if (!isFresh(roleID)) {
                    return this.SystemAPI.roleDelete({ roleID });
                }
            });
        });
    }
    /**
     * Assign role to user
     *
     * @example
     * addUserToRole('user-we-can-trust', 'admins')
     *
     * @param user resolvable user input
     * @param role resolvable role input
     */
    addUserToRole(user, role) {
        return __awaiter(this, void 0, void 0, function* () {
            let userID;
            let roleID;
            return this.resolveUser(user, this.$user).then(user => {
                userID = extractID(user, 'userID');
                return this.resolveRole(role, this.$role);
            }).then(role => {
                roleID = extractID(role, 'roleID');
                return this.SystemAPI.roleMemberAdd({ roleID, userID });
            });
        });
    }
    /**
     * Remove role from user
     * @example
     * addUserToRole('user-we-can-trust', 'admins')
     *
     * @param user - resolvable user input
     * @param role - resolvable role input
     */
    removeUserFromRole(user, role) {
        return __awaiter(this, void 0, void 0, function* () {
            let userID;
            let roleID;
            return this.resolveUser(user, this.$user).then(user => {
                userID = extractID(user, 'userID');
                return this.resolveRole(role, this.$role);
            }).then(role => {
                roleID = extractID(role, 'roleID');
                return this.SystemAPI.roleMemberRemove({ roleID, userID });
            });
        });
    }
    /**
     * Resolves users from the arguments and returns first valid
     *
     * Knows how to resolve from:
     *  - string that looks like an ID - find by id (fallback to find-by-handle)
     *  - string that looks like an email - find by email (fallback to find-by-handle)
     *  - string - find by handle
     *  - User object
     *  - object with userID or ownerID properties
     */
    resolveUser(...args) {
        return __awaiter(this, void 0, void 0, function* () {
            for (let u of args) {
                // Resolve pending promises if any...
                u = yield u;
                if (!u) {
                    continue;
                }
                if (typeof u === 'string') {
                    try {
                        if (IsCortezaID(u)) {
                            // Looks like an ID, try to find it and fall back to handle
                            return yield this.findUserByID(u);
                        }
                        else if (u.indexOf('@') > 0) {
                            return yield this.findUserByEmail(u);
                        }
                    }
                    catch (e) {
                        console.error(e);
                    }
                    // Always fall back to handle
                    return this.findUserByHandle(u);
                }
                if (typeof u !== 'object') {
                    continue;
                }
                if (isUser(u)) {
                    // Already got what we need
                    return Promise.resolve(u);
                }
                // Other kind of object with properties that might hold user ID
                const { userID, ownerID, } = u;
                return this.resolveUser(userID, ownerID);
            }
            return Promise.reject(new Error('unexpected input type for user resolver'));
        });
    }
    /**
     * Resolves users from the arguments and returns first valid
     *
     * Knows how to resolve from:
     *  - string that looks like an ID - find by id (fallback to find-by-handle)
     *  - string - find by handle
     *  - Role object
     *  - object with roleID property
     */
    resolveRole(...args) {
        return __awaiter(this, void 0, void 0, function* () {
            for (let r of args) {
                // Resolve pending promises if any...
                r = yield r;
                if (!r) {
                    continue;
                }
                if (typeof r === 'string') {
                    if (IsCortezaID(r)) {
                        // Looks like an ID, try to find it and fall back to handle
                        return this.findRoleByID(r).catch(() => this.findRoleByHandle(r));
                    }
                    return this.findRoleByHandle(r);
                }
                if (typeof r !== 'object') {
                    continue;
                }
                if (isRole(r)) {
                    // Already got what we need
                    return r;
                }
                // Other kind of object with properties that might hold role ID
                const { roleID, } = r;
                return this.resolveRole(roleID);
            }
            return Promise.reject(Error('unexpected input type for role resolver'));
        });
    }
    /**
     * Allows access for the given role for the given System resource
     *
     * @example
     * // Allows users with `someRole` to access the newly created user
     * await Compose.allow({
     *    role: someRole,
     *    resource: newUser,
     *    operation: 'read',
     * })
     */
    allow(...pr) {
        return __awaiter(this, void 0, void 0, function* () {
            const rr = pr.map(p => ({
                role: p.role,
                resource: p.resource,
                operation: p.operation,
                access: 'allow',
            }));
            return genericPermissionUpdater(this.SystemAPI, rr);
        });
    }
    /**
     * Denies access for the given role for the given System resource
     *
     * @example
     * // Denies users with `someRole` from accessing the newly created user
     * await Compose.deny({
     *    role: someRole,
     *    resource: newUser,
     *    operation: 'read',
     * })
     */
    deny(...pr) {
        return __awaiter(this, void 0, void 0, function* () {
            const rr = pr.map(p => ({
                role: p.role,
                resource: p.resource,
                operation: p.operation,
                access: 'deny',
            }));
            return genericPermissionUpdater(this.SystemAPI, rr);
        });
    }
    /**
     * Inherits access for the given role for the given System resource
     *
     * @example
     * // Uses inherited permissions for the `sameRole` for the newly created user
     * await Compose.inherit({
     *    role: someRole,
     *    resource: newUser,
     *    operation: 'read',
     * })
     */
    inherit(...pr) {
        return __awaiter(this, void 0, void 0, function* () {
            const rr = pr.map(p => ({
                role: p.role,
                resource: p.resource,
                operation: p.operation,
                access: 'inherit',
            }));
            return genericPermissionUpdater(this.SystemAPI, rr);
        });
    }
}

const { merge: merge$9 } = lodash;
const FieldNameValidator = /^[A-Za-z][0-9A-Za-z_-]*[A-Za-z0-9]$/;
const unsortableFieldKinds = ['File', 'Geometry'];
const unsortableSysFields = ['recordID'];
const unfilterableFieldKinds = ['File', 'Geometry'];
const nonQueryableFieldKinds = ['Number', 'Record', 'User', 'Bool', 'DateTime', 'File', 'Geometry'];
const nonQueryableFieldNames = ['recordID'];
const defaultOptions = () => Object.freeze({
    description: {
        view: '',
        edit: undefined,
    },
    hint: {
        view: '',
        edit: undefined,
    },
});
class ModuleField {
    constructor(f) {
        this.fieldID = NoID;
        this.name = '';
        this.kind = '';
        this.label = '';
        this.defaultValue = [];
        this.maxLength = 0;
        this.isRequired = false;
        this.isMulti = false;
        this.isSystem = false;
        this.isSortable = true;
        this.isFilterable = true;
        this.isQueryable = true;
        this.options = Object.assign({}, defaultOptions());
        this.expressions = {};
        this.config = {
            dal: {
                encodingStrategy: null,
            },
            privacy: {
                sensitivityLevelID: NoID,
                usageDisclosure: '',
            },
            recordRevisions: {
                enabled: false,
            },
        };
        this.canUpdateRecordValue = false;
        this.canReadRecordValue = false;
        this.apply(f);
    }
    applyOptions(o) {
        if (!o)
            return;
        if (o.description) {
            this.options.description = Object.assign(Object.assign({}, this.options.description), o.description);
            this.options.description.edit = this.options.description.edit || undefined;
        }
        if (o.hint) {
            this.options.hint = Object.assign(Object.assign({}, this.options.hint), o.hint);
            this.options.hint.edit = this.options.hint.edit || undefined;
        }
    }
    clone() {
        return new ModuleField(JSON.parse(JSON.stringify(this)));
    }
    apply(f) {
        if (!f)
            return;
        Apply(this, f, CortezaID, 'fieldID');
        Apply(this, f, String, 'name', 'label', 'kind');
        Apply(this, f, Number, 'maxLength');
        Apply(this, f, Boolean, 'isRequired', 'isMulti', 'isSystem');
        // Make sure field is align with it's capabilities
        if (!this.cap.multi)
            this.isMulti = false;
        if (!this.cap.required)
            this.isRequired = false;
        // Check if kind sortable
        if (unsortableFieldKinds.includes(this.kind)) {
            this.isSortable = false;
        }
        if (unfilterableFieldKinds.includes(this.kind)) {
            this.isFilterable = false;
        }
        if (nonQueryableFieldKinds.includes(this.kind)) {
            this.isQueryable = false;
        }
        if (nonQueryableFieldNames.includes(this.name)) {
            this.isQueryable = false;
        }
        if (f.defaultValue && Array.isArray(f.defaultValue)) {
            /**
             * Converting default value into proper format
             * so we can use it without conversion
             * false boolean values are represented only by the name, in all other cases the value is also present
             */
            this.defaultValue = f.defaultValue.filter(({ name, value }) => name !== undefined || (value !== undefined && value !== null));
        }
        if (this.isSystem) {
            this.canUpdateRecordValue = true;
            this.canReadRecordValue = true;
            if (unsortableSysFields.includes(this.name)) {
                this.isSortable = false;
            }
        }
        else {
            Apply(this, f, Boolean, 'canUpdateRecordValue', 'canReadRecordValue');
        }
        if (IsOf(f, 'config')) {
            this.config = merge$9({}, this.config, f.config);
        }
        if (IsOf(f, 'kind')) {
            this.kind = f.kind;
        }
        if (IsOf(f, 'expressions')) {
            this.expressions = f.expressions;
        }
    }
    /**
     * Test field validity
     *
     * Expecting valid name
     */
    get isValid() {
        return this.name.length > 0 && FieldNameValidator.test(this.name);
    }
    /**
     * Per module field type capabilities
     */
    get cap() {
        return {
            configurable: true,
            multi: true,
            writable: true,
            required: true,
            private: true,
        };
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.fieldID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'compose:module-field';
    }
}
const Registry$1 = new Map();

const kind$s = 'Bool';
const defaults$s = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { trueLabel: '', falseLabel: '', switch: false }));
class ModuleFieldBool extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$s;
        this.options = Object.assign({}, defaults$s());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, String, 'trueLabel', 'falseLabel');
        Apply(this.options, o, Boolean, 'switch');
    }
    /**
     * Per module field type capabilities
     */
    get cap() {
        return Object.assign(Object.assign({}, super.cap), { multi: false });
    }
}
Registry$1.set(kind$s, ModuleFieldBool);

/**
 * Returns current language
 *
 * This temporary solution returns an empty array;
 * this will cause Intl functions to format strings and numbers in the current (by-browser) language
 */
function currentLanguage() {
    var _a;
    return navigator.language || ((_a = navigator.languages) === null || _a === void 0 ? void 0 : _a[0]) || 'en-US';
}

/**
 * Parses input into Date using Moment library
 *
 * @param input
 */
function parse(input) {
    return moment(input).toDate();
}
function format(input, options) {
    return (new Intl.DateTimeFormat(currentLanguage(), options)).format(parse(input));
}
/**
 * Outputs locally formatted date and time, no seconds
 *
 * Examples:
 * "Wednesday, September 8, 2021 at 9:41 AM"
 * "sreda, 08. september 2021 09:41"
 * "srijeda, 8. rujna 2021. u 09:42"
 *
 * @param input
 * @param options
 */
function fullDateTime(input, options = { dateStyle: 'full', timeStyle: 'short' }) {
    return format(input, options);
}
/**
 * Outputs locally formatted date without time
 *
 * Example:
 * 09/04/1986
 *
 * @param input
 * @param options
 */
function date(input, options = { dateStyle: 'short' }) {
    return format(input, options);
}
/**
 * Outputs locally formatted time
 *
 * Example:
 * 8:30 PM
 *
 * @param input
 * @param options
 */
function time(input, options = { timeStyle: 'short' }) {
    return format(input, options);
}

function number(input, options = { maximumFractionDigits: 6 }) {
    return new Intl.NumberFormat(currentLanguage(), options).format(input);
}
function accountingNumber(value) {
    if (value === 0) {
        return '-';
    }
    if (value < 0) {
        return `(${number(Math.abs(value))})`;
    }
    return number(value);
}

var index$4 = /*#__PURE__*/Object.freeze({
  __proto__: null,
  accountingNumber: accountingNumber,
  date: date,
  fullDateTime: fullDateTime,
  number: number,
  time: time
});

// @todo option to allow only time entry
// @todo option to allow multiple entries
// @todo option to allow duplicates
const kind$r = 'DateTime';
const defaults$r = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { format: '', multiDelimiter: '\n', onlyDate: false, onlyFutureValues: false, onlyPastValues: false, onlyTime: false, outputRelative: false }));
class ModuleFieldDateTime extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$r;
        this.options = Object.assign({}, defaults$r());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, String, 'format', 'multiDelimiter');
        Apply(this.options, o, Boolean, 'onlyDate', 'onlyTime', 'onlyPastValues', 'onlyFutureValues', 'outputRelative');
    }
    formatValue(value) {
        if (value === 'Invalid date') {
            return null;
        }
        const o = this.options;
        const m = moment(value, ['YYYY-MM-DDTHH:mm:ssZ', 'YYYY-MM-DD HH:mm', 'YYYY-MM-DD', 'HH:mm']);
        if (o.outputRelative) {
            return m.fromNow();
        }
        else if (o.format.length > 0) {
            return m.format(o.format);
        }
        else if (o.onlyTime) {
            return time(m);
        }
        else if (o.onlyDate) {
            return date(moment(value, 'YYYY-MM-DD'));
        }
        else {
            return fullDateTime(m);
        }
    }
    /**
     * Checks if given value is in the future
     * @param {String|Array<String>} v Value (in DateTime) to check
     * @param {Moment} now Time reference
     * @returns {undefined|String} undefined if valid, Error string if invalid
     */
    checkFuture(v, now = moment()) {
        if (!this.options.onlyFutureValues) {
            return undefined;
        }
        if (!Array.isArray(v)) {
            v = [v];
        }
        if (v.find(v => moment(v) < now)) {
            return 'notification.field-datetime.valueNotFuture';
        }
        return undefined;
    }
    /**
     * Checks if given value is in the past
     * @param {String|Array<String>} v Value (in DateTime) to check
     * @param {Moment} now Time reference
     * @returns {undefined|String} undefined if valid, Error string if invalid
     */
    checkPast(v, now = moment()) {
        if (!this.options.onlyPastValues) {
            return undefined;
        }
        if (!Array.isArray(v)) {
            v = [v];
        }
        if (v.find(v => moment(v) > now)) {
            return 'notification.field-datetime.valueNotPast';
        }
    }
    /**
     * Checks if given value is valid for this field
     * @param {String} v Value (in DateTime) to check
     * @param {Moment} now Reference time used to compare
     * @returns {Array<>} Array of issues; empty if none
     */
    validate(v, now = moment()) {
        let err = this.checkFuture(v, now);
        err = err || this.checkPast(v, now);
        if (err) {
            return [err];
        }
        return [];
    }
}
Registry$1.set(kind$r, ModuleFieldDateTime);

// @todo option to allow multiple entries
// @todo option to allow duplicates
// @todo option to allow only whitelisted domains
const kind$q = 'Email';
const defaults$q = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { outputPlain: true, multiDelimiter: '\n' }));
class ModuleFieldEmail extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$q;
        this.options = Object.assign({}, defaults$q());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, String, 'multiDelimiter');
        Apply(this.options, o, Boolean, 'outputPlain');
    }
}
Registry$1.set(kind$q, ModuleFieldEmail);

const kind$p = 'File';
const modes = [
    // list of attachments, no preview
    'list',
    // list of all images/files, show preview
    'gallery',
];
const defaults$p = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { allowImages: true, allowDocuments: true, maxSize: 0, mode: 'list', inline: true, hideFileName: false, mimetypes: '', height: '', width: '', maxHeight: '', maxWidth: '', borderRadius: '', margin: 'auto', backgroundColor: '#FFFFFF00', clickToView: true, enableDownload: true, multiDelimiter: '\n', enableWebcam: false }));
class ModuleFieldFile extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$p;
        this.options = Object.assign({}, defaults$p());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, Number, 'maxSize');
        Apply(this.options, o, Boolean, 'allowImages', 'allowDocuments', 'inline', 'hideFileName', 'clickToView', 'enableDownload', 'enableWebcam');
        Apply(this.options, o, String, 'mimetypes', 'height', 'width', 'maxHeight', 'maxWidth', 'borderRadius', 'margin', 'backgroundColor');
        // Legacy
        if (o.mode === 'single') {
            o.mode = 'gallery';
        }
        else if (o.mode === 'grid') {
            o.mode = 'list';
        }
        ApplyWhitelisted(this.options, o, modes, 'mode');
    }
}
Registry$1.set(kind$p, ModuleFieldFile);

const kind$o = 'Select';
const defaults$o = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { options: [], selectType: 'default', multiDelimiter: '\n', isUniqueMultiValue: false, displayType: 'text' }));
class ModuleFieldSelect extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$o;
        this.options = Object.assign({}, defaults$o());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, String, 'selectType', 'multiDelimiter', 'displayType');
        Apply(this.options, o, Boolean, 'isUniqueMultiValue');
        if (o.options) {
            let opt = [];
            if (AreStrings(o.options)) {
                opt = o.options.map((value) => this.createSelectOption({ value, text: value }));
            }
            else {
                opt = o.options.map(o => this.createSelectOption(o));
            }
            this.options.options = opt;
        }
    }
    createSelectOption({ value = '', text = '', style = {} } = {}) {
        const { textColor = '', backgroundColor = '' } = style || {};
        return {
            value,
            text,
            style: {
                textColor,
                backgroundColor,
            },
        };
    }
}
Registry$1.set(kind$o, ModuleFieldSelect);

const kind$n = 'Number';
const defaults$n = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { presetFormat: 'custom', precision: 3, multiDelimiter: '\n', display: 'number', 
    // Number display options
    format: '', prefix: '', suffix: '', 
    // Progress bar display options
    min: 0, max: 100, step: 1, showValue: true, showRelative: true, showProgress: false, animated: false, variant: 'success', thresholds: [] }));
class ModuleFieldNumber extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$n;
        this.options = Object.assign({}, defaults$n());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, String, 'format', 'prefix', 'suffix', 'multiDelimiter', 'display', 'variant', 'presetFormat');
        Apply(this.options, o, Number, 'precision', 'min', 'max', 'step');
        Apply(this.options, o, Boolean, 'showValue', 'showRelative', 'showProgress', 'animated');
        if (o.thresholds) {
            this.options.thresholds = o.thresholds;
        }
    }
    formatValue(value, format) {
        const o = this.options;
        let n;
        format = o.presetFormat === 'custom' ? o.format : o.presetFormat;
        switch (typeof value) {
            case 'string':
                n = parseFloat(value);
                break;
            case 'number':
                n = value;
                break;
            default:
                n = 0;
        }
        let out = `${n}`;
        if (format === 'accounting') {
            out = accountingNumber(n);
        }
        else if (format && format.length > 0) {
            out = numeral(n).format(format);
        }
        else {
            out = number(n);
        }
        return '' + o.prefix + (out || n) + o.suffix;
    }
}
Registry$1.set(kind$n, ModuleFieldNumber);

// @todo option to allow multiple entries
// @todo option to allow duplicates
const kind$m = 'Record';
const defaults$m = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { moduleID: NoID, labelField: '', recordLabelField: '', queryFields: [], selectType: '', multiDelimiter: '\n', isUniqueMultiValue: false, prefilter: undefined }));
class ModuleFieldRecord extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$m;
        this.options = Object.assign({}, defaults$m());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, CortezaID, 'moduleID');
        Apply(this.options, o, String, 'labelField', 'recordLabelField', 'selectType', 'multiDelimiter', 'prefilter');
        Apply(this.options, o, Boolean, 'isUniqueMultiValue');
        Apply(this.options, o, (o) => {
            if (!o) {
                return [];
            }
            if (!Array.isArray(o)) {
                return [o];
            }
            return o;
        }, 'queryFields');
    }
}
Registry$1.set(kind$m, ModuleFieldRecord);

const kind$l = 'String';
const defaults$l = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { multiLine: false, useRichTextEditor: false, multiDelimiter: '\n' }));
class ModuleFieldString extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$l;
        this.options = Object.assign({}, defaults$l());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, String, 'multiDelimiter');
        Apply(this.options, o, Boolean, 'multiLine', 'useRichTextEditor');
    }
}
Registry$1.set(kind$l, ModuleFieldString);

// @todo option to allow multiple entries
// @todo option to allow duplicates
const kind$k = 'Url';
const defaults$k = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { trimFragment: false, trimQuery: false, trimPath: false, onlySecure: false, outputPlain: false, multiDelimiter: '\n' }));
class ModuleFieldUrl extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$k;
        this.options = Object.assign({}, defaults$k());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, String, 'multiDelimiter');
        Apply(this.options, o, Boolean, 'trimFragment', 'trimQuery', 'trimPath', 'onlySecure', 'outputPlain');
    }
}
Registry$1.set(kind$k, ModuleFieldUrl);

// @todo option to allow multiple entries
// @todo option to allow duplicates
const kind$j = 'User';
const defaults$j = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { roles: [], presetWithAuthenticated: false, selectType: 'default', multiDelimiter: '\n', isUniqueMultiValue: false }));
class ModuleFieldUser extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$j;
        this.options = Object.assign({}, defaults$j());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, Boolean, 'presetWithAuthenticated', 'isUniqueMultiValue');
        Apply(this.options, o, String, 'selectType', 'multiDelimiter');
        Apply(this.options, o, (o) => {
            if (!o) {
                return [];
            }
            if (!Array.isArray(o)) {
                return [o];
            }
            return o;
        }, 'roles');
    }
    formatter({ userID, name, username, email, handle } = {}) {
        return name || username || email || handle || userID || '';
    }
}
Registry$1.set(kind$j, ModuleFieldUser);

const kind$i = 'Geometry';
const defaults$i = () => Object.freeze(Object.assign(Object.assign({}, defaultOptions()), { center: [30, 30], zoom: 3, multiDelimiter: '\n', prefillWithCurrentLocation: false, hideCurrentLocationButton: false, hideGeoSearch: false }));
class ModuleFieldGeometry extends ModuleField {
    constructor(i) {
        super(i);
        this.kind = kind$i;
        this.options = Object.assign({}, defaults$i());
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        super.applyOptions(o);
        Apply(this.options, o, String, 'multiDelimiter');
        Apply(this.options, o, Number, 'zoom');
        Apply(this.options, o, Boolean, 'prefillWithCurrentLocation', 'hideCurrentLocationButton', 'hideGeoSearch');
        if (o.center) {
            this.options.center = o.center;
        }
    }
    /**
     * Per module field type capabilities
     */
    get cap() {
        return Object.assign(Object.assign({}, super.cap), { multi: true });
    }
}
Registry$1.set(kind$i, ModuleFieldGeometry);

function ModuleFieldMaker(i) {
    if (!i.kind) {
        return new ModuleField(i);
    }
    if (!Registry$1.has(i.kind)) {
        throw new Error(`unknown module field kind '${i.kind}'`);
    }
    return new (Registry$1.get(i.kind))(i);
}

class Namespace {
    constructor(i) {
        this.namespaceID = NoID;
        this.name = '';
        this.slug = '';
        this.enabled = false;
        this.labels = {};
        this.meta = {};
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.canCreateChart = false;
        this.canCreateModule = false;
        this.canCreatePage = false;
        this.canDeleteNamespace = false;
        this.canUpdateNamespace = false;
        this.canManageNamespace = false;
        this.canCloneNamespace = false;
        this.canExportNamespace = false;
        this.canGrant = false;
        this.canExportCharts = false;
        this.canExportModules = false;
        this.apply(i);
    }
    clone() {
        return new Namespace(JSON.parse(JSON.stringify(this)));
    }
    apply(n) {
        if (!n)
            return;
        Apply(this, n, CortezaID, 'namespaceID');
        Apply(this, n, String, 'name', 'slug');
        Apply(this, n, Boolean, 'enabled');
        if (IsOf(n, 'meta')) {
            this.meta = Object.assign({}, n.meta);
        }
        if (IsOf(n, 'labels')) {
            this.labels = Object.assign({}, n.labels);
        }
        Apply(this, n, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, n, Boolean, 'canDeleteNamespace', 'canUpdateNamespace', 'canManageNamespace', 'canCloneNamespace', 'canExportNamespace', 'canGrant', 'canCreateModule', 'canExportModules', 'canCreatePage', 'canCreateChart', 'canExportCharts');
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.namespaceID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'compose:namespace';
    }
    /**
     * Calculate namespace initials
     */
    get initials() {
        let base = this.name || this.slug;
        // if length is shorter than 3 letters, use that
        if (base.length <= 3) {
            return base;
        }
        // split by space and take first letter of each word
        base = base.split(/\s+/).map(w => w[0]).filter(c => /[a-zA-Z]/.test(c)).join('');
        if (base.length > 3) {
            base = base.slice(0, 3);
        }
        return base;
    }
}

const { merge: merge$8 } = lodash;
const propNamespace = Symbol('namespace');
/**
 * System fields that are present in every record.
 */
const systemFields = Object.freeze([
    { isSystem: true, name: 'recordID', label: 'Record ID', kind: 'String' },
    { isSystem: true, name: 'ownedBy', label: 'Owned by', kind: 'User' },
    { isSystem: true, name: 'createdBy', label: 'Created by', kind: 'User' },
    { isSystem: true, name: 'createdAt', label: 'Created at', kind: 'DateTime' },
    { isSystem: true, name: 'updatedBy', label: 'Updated by', kind: 'User' },
    { isSystem: true, name: 'updatedAt', label: 'Updated at', kind: 'DateTime' },
    { isSystem: true, name: 'revision', label: 'Revision', kind: 'Number' },
    { isSystem: true, name: 'deletedBy', label: 'Deleted by', kind: 'User' },
    { isSystem: true, name: 'deletedAt', label: 'Deleted at', kind: 'DateTime' },
].map(f => ModuleFieldMaker(f)));
class Module {
    constructor(i, ns) {
        this.moduleID = NoID;
        this.namespaceID = NoID;
        this.name = '';
        this.handle = '';
        this.fields = [];
        this.issues = [];
        this.config = {
            dal: {
                connectionID: NoID,
                ident: '',
                systemFieldEncoding: {
                    id: null,
                    revision: null,
                    moduleID: null,
                    namespaceID: null,
                    ownedBy: null,
                    createdBy: null,
                    createdAt: null,
                    updatedBy: null,
                    updatedAt: null,
                    deletedBy: null,
                    deletedAt: null,
                },
            },
            privacy: {
                sensitivityLevelID: NoID,
                usageDisclosure: '',
            },
            discovery: {
                public: {
                    result: [
                        {
                            lang: '',
                            fields: [],
                        },
                    ],
                },
                private: {
                    result: [
                        {
                            lang: '',
                            fields: [],
                        },
                    ],
                },
                protected: {
                    result: [
                        {
                            lang: '',
                            fields: [],
                        },
                    ],
                },
            },
            recordRevisions: {
                enabled: false,
                ident: '',
            },
            recordDeDup: {
                rules: [],
            },
        };
        this.meta = {
            ui: {
                admin: {
                    fields: [],
                },
            },
        };
        this.labels = {};
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.canUpdateModule = false;
        this.canDeleteModule = false;
        this.canCreateRecord = false;
        this.canCreateOwnedRecord = false;
        this.canGrant = false;
        if (ns) {
            this.namespace = ns;
        }
        this.apply(i);
    }
    clone() {
        return new Module(JSON.parse(JSON.stringify(this)), this.namespace);
    }
    apply(m) {
        if (!m)
            return;
        if (this.namespace && m.namespaceID && m.namespaceID !== this.namespace.namespaceID) {
            throw new Error('module can not change namespace');
        }
        Apply(this, m, CortezaID, 'moduleID', 'namespaceID');
        Apply(this, m, String, 'name', 'handle');
        if (IsOf(m, 'fields')) {
            this.fields = [];
            if (AreObjects(m.fields)) {
                // We're very permissive here -- array of (empty) objects is all we need
                // to create fields.
                this.fields = m.fields.map((b) => ModuleFieldMaker(b));
            }
        }
        if (IsOf(m, 'meta')) {
            if (m.meta.ui && m.meta.ui.admin && m.meta.ui.admin.fields) {
                if (!AreStrings(m.meta.ui.admin.fields)) {
                    const fields = m.meta.ui.admin.fields || [];
                    m.meta.ui.admin.fields = fields.map((f) => f.fieldID && f.fieldID !== NoID ? f.fieldID : f.name).filter((f) => !!f);
                }
            }
            this.meta = merge$8({}, this.meta, m.meta);
        }
        if (IsOf(m, 'config')) {
            this.config = merge$8({}, this.config, m.config);
        }
        if (IsOf(m, 'labels')) {
            this.labels = Object.assign({}, m.labels);
        }
        if (IsOf(m, 'issues')) {
            this.issues = m.issues;
        }
        Apply(this, m, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, m, Boolean, 'canUpdateModule', 'canDeleteModule', 'canCreateRecord', 'canCreateOwnedRecord', 'canGrant');
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.moduleID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'compose:module';
    }
    get namespace() {
        return this[propNamespace];
    }
    set namespace(ns) {
        if (this[propNamespace]) {
            if (this[propNamespace].namespaceID !== ns.namespaceID) {
                throw new Error('namespace for this module already set');
            }
        }
        this.namespaceID = ns.namespaceID;
        if (Object.isFrozen(ns)) {
            this[propNamespace] = ns;
        }
        else {
            // Making a copy and freezing it
            this[propNamespace] = Object.freeze(new Namespace(ns));
        }
        this[propNamespace] = ns;
    }
    /**
     * Returns fields from module, filtered and order as requested
     */
    filterFields(requested) {
        if (!requested || requested.length === 0) {
            return [];
        }
        if (!AreStrings(requested)) {
            requested = requested.map((f) => f.name || f.fieldID);
        }
        const out = [];
        for (const r of requested) {
            const sf = this.systemFields().find(f => r === f.name || r === f.fieldID);
            if (sf) {
                out.push(sf);
                continue;
            }
            const mf = this.fields.find(f => r === f.name || r === f.fieldID);
            if (mf) {
                out.push(mf);
            }
        }
        return out;
    }
    findField(name) {
        const r = this.filterFields([name]);
        return r && r.length > 0 ? r[0] : undefined;
    }
    fieldNames() {
        return this.fields.map(f => f.name);
    }
    systemFields() {
        return systemFields;
    }
    export() {
        return this;
    }
    import() {
        return this;
    }
}

var _a;
const fieldIndex = Symbol('fieldIndex');
const propModule = Symbol('module');
const cleanValues = Symbol('cleanValues');
const reservedFieldNames = [
    'toJSON',
];
/**
 * For something to be useful module (for a Record), it needs to contain fields
 */
function isModule$1(m) {
    return !!m && IsOf(m, 'fields') && Array.isArray(m.fields) && m.fields.length > 0;
}
/**
 * Record class will be used all over the place, user scripts, etc..
 *
 * Constructor (and apply fn) is as versatile as possible to handle
 * different use-cases.
 */
class Record {
    constructor(recModVal1, recModVal2) {
        this.recordID = NoID;
        this.moduleID = NoID;
        this.namespaceID = NoID;
        this.revision = 0;
        this.values = {};
        this.valueErrors = {};
        this.meta = {};
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.ownedBy = undefined;
        this.createdBy = undefined;
        this.updatedBy = undefined;
        this.deletedBy = undefined;
        this.canUpdateRecord = false;
        this.canReadRecord = false;
        this.canDeleteRecord = false;
        this.canUndeleteRecord = false;
        this.canManageOwnerOnRecord = false;
        this.canSearchRevision = false;
        this.canGrant = false;
        this[_a] = {};
        if (recModVal1 instanceof Record) {
            this.module = recModVal1.module;
            this.apply(recModVal1);
            return;
        }
        if (isModule$1(recModVal1)) {
            this.module = recModVal1;
            this.apply(recModVal2);
            return;
        }
        if (isModule$1(recModVal2)) {
            this.module = recModVal2;
            this.apply(recModVal1);
            return;
        }
        throw new Error('invalid module used to initialize a record');
    }
    clone() {
        return new Record(this.module, JSON.parse(JSON.stringify(this)));
    }
    /**
     * apply (partially) updates record and it's values
     *
     * @param p
     */
    apply(p) {
        if (p === undefined) {
            // This is a brand new record; set default values
            this.defaultValues();
            return;
        }
        let r;
        // Determine what kind of value we got
        switch (true) {
            case IsOf(p, 'recordID') || IsOf(p, 'values'):
                // p1 is something that looks like a record object
                r = p;
                break;
            case AreObjectsOf(p, 'name'):
                // assuming p1 is array of raw values
                r = ({ values: p });
                break;
            default:
                r = ({ values: p });
        }
        r = r;
        if (this.module && r.moduleID && r.moduleID !== this.module.moduleID) {
            throw new Error('can not change module on a record');
        }
        if (this.namespace && r.namespaceID && r.namespaceID !== this.namespace.namespaceID) {
            throw new Error('can not change namespace on a record');
        }
        if (r.namespaceID && r.namespaceID !== this.module.namespaceID) {
            throw new Error('record and module namespace do not match');
        }
        Apply(this, r, CortezaID, 'recordID', 'moduleID', 'namespaceID');
        Apply(this, r, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, r, CortezaID, 'ownedBy', 'createdBy', 'updatedBy', 'deletedBy');
        Apply(this, r, Number, 'revision');
        Apply(this, r, Boolean, 'canUpdateRecord', 'canReadRecord', 'canDeleteRecord', 'canUndeleteRecord', 'canManageOwnerOnRecord', 'canGrant');
        // This is a brand-new record; set default values
        if (!r.recordID || r.recordID === NoID) {
            this.defaultValues();
        }
        if (r.values !== undefined) {
            this.updateValues(r.values);
        }
        if (!this[cleanValues]) {
            // When there are no clean values,
            // make copy of values so that we know if change occurred
            this[cleanValues] = Object.freeze(Object.assign({}, this.values));
        }
        if (r.valueErrors) {
            this.valueErrors = r.valueErrors;
        }
        if (IsOf(r, 'meta')) {
            this.meta = Object.assign({}, r.meta);
        }
    }
    get cleanValues() {
        return this[cleanValues];
    }
    get module() {
        if (this[propModule] === undefined) {
            throw new Error('module not set');
        }
        return this[propModule];
    }
    set module(m) {
        if (this[propModule]) {
            if (this[propModule].moduleID !== m.moduleID) {
                throw new Error('module for this record already set');
            }
        }
        if (!m.fields || !Array.isArray(m.fields) || m.fields.length === 0) {
            throw new Error('module used to initialize a record does not contain any fields');
        }
        this.moduleID = m.moduleID;
        this.namespaceID = m.namespaceID;
        this[fieldIndex] = new Map();
        if (Object.isFrozen(m)) {
            this[propModule] = m;
        }
        else {
            // Making a copy and freezing it
            this[propModule] = Object.freeze(new Module(m));
        }
        this[propModule].fields.forEach(f => {
            const { name, isMulti, kind, defaultValue, } = f;
            if (reservedFieldNames.includes(name)) {
                throw new Error('can not use reserved field name ' + name);
            }
            this[fieldIndex].set(name, { isMulti, kind, defaultValue });
        });
        Object.freeze(this[fieldIndex]);
        this.initValues();
    }
    get namespace() {
        return this.module.namespace;
    }
    /**
     * Converts internal representation of values into array of RawValue objects
     */
    serializeValues() {
        const vv = [];
        this[fieldIndex].forEach(({ isMulti }, name) => {
            if (this.values[name] === undefined) {
                return;
            }
            const val = this.values[name];
            if (isMulti) {
                if (Array.isArray(this.values[name])) {
                    for (let i = 0; i < val.length; i++) {
                        if (val[i] !== undefined) {
                            vv.push({ name, value: val[i].toString() });
                        }
                    }
                }
            }
            else {
                vv.push({ name, value: val.toString() });
            }
        });
        return vv;
    }
    /**
     * Removes existing, resets default values and updates it with new ones
     */
    setValues(...i) {
        this.initValues();
        this.defaultValues();
        this.updateValues(...i);
    }
    /**
     * Removes existing and resets default values
     */
    initValues() {
        const dst = {};
        this[fieldIndex].forEach(({ isMulti }, name) => {
            if (isMulti) {
                dst[name] = [];
            }
            else {
                dst[name] = undefined;
            }
        });
        // TypeScript complains about incompatibility between
        // indexed object and toJSON function
        // @ts-ignore
        dst.toJSON = () => this.serializeValues();
        this.values = dst;
    }
    defaultValues() {
        this[fieldIndex].forEach(({ isMulti, defaultValue }, name) => {
            if (defaultValue && Array.isArray(defaultValue) && defaultValue.length > 0) {
                if (isMulti) {
                    this.values[name] = defaultValue.map(({ value }) => value);
                }
                else {
                    this.values[name] = defaultValue[0].value;
                }
            }
        });
    }
    /**
     * Updates record's values object with provided input
     *
     * Accepted values:
     * 1. Array of RawValue objects:
     *    updateValues([{ name: ..., value: ...}, ...])
     *
     * 2. One or more Value object:
     *    updateValues({ foo: ..., bar: ... }, ...)
     */
    updateValues(...combo) {
        // If all values are formatted as raw value
        if (combo.length === 1 && AreObjectsOf(combo[0], 'name')) {
            combo[0].forEach(({ name, value }) => this.setValue(name, value));
            return;
        }
        combo.forEach(v => {
            if (Array.isArray(v)) {
                this.updateValues(...v);
                return;
            }
            if (!v || typeof v !== 'object') {
                throw Error('expecting array of values or values object');
            }
            // Handle Values
            for (const name of Object.getOwnPropertyNames(v)) {
                this.setValue(name, v[name]);
            }
        });
    }
    /**
     * Sets single value
     *
     * @param name
     * @param value
     */
    setValue(name, value, index = -1) {
        // Skip reserved names
        if (reservedFieldNames.includes(name)) {
            return;
        }
        // Skip unknown fields
        if (!this[fieldIndex].has(name)) {
            return;
        }
        const { kind, isMulti } = this[fieldIndex].get(name);
        if (value === undefined || value.length === 0) {
            // nothing given, nothing set
            this.values[name] = isMulti ? [] : (kind === 'Bool' ? '0' : undefined);
            return;
        }
        if (isMulti) {
            if (Array.isArray(value)) {
                if (index < -1) {
                    // assigning [] to [i]
                    throw Error('can not set array of values to a single value');
                }
                this.values[name] = Array.isArray(value) ? value : [value];
                return;
            }
            if (index === -1) {
                this.values[name].push(value);
                return;
            }
            this.values[name][index] = value;
            return;
        }
        if (Array.isArray(value)) {
            value = value[0];
        }
        // Update with first item or set to undefined
        this.values[name] = value;
    }
    serialize() {
        const _b = this.values, { toJSON } = _b, values = __rest(_b, ["toJSON"]);
        return Object.assign(Object.assign({}, this), { values });
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.recordID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'compose:record';
    }
    /**
     * Proxy to Record's meta to maintain BC
     */
    get labels() {
        return this.meta;
    }
    get properties() {
        return [
            'recordID',
            'moduleID',
            'namespaceID',
            'revision',
            'meta',
            'createdAt',
            'updatedAt',
            'deletedAt',
            'ownedBy',
            'createdBy',
            'updatedBy',
            'deletedBy',
            'canUpdateRecord',
            'canReadRecord',
            'canDeleteRecord',
            'canUndeleteRecord',
            'canManageOwnerOnRecord',
            'canSearchRevision',
            'canGrant',
        ];
    }
}
_a = cleanValues;

function isRawRevisionPayload(raw) {
    if (!raw || typeof raw !== 'object') {
        console.warn('not an object', raw);
        return false;
    }
    if (!Object.getOwnPropertyNames(raw).includes('set')) {
        console.warn('no set prop', raw);
        return false;
    }
    if (!Array.isArray(raw.set)) {
        console.warn('set prop not array', raw);
        return false;
    }
    return true;
}
function convertRevisionPayloadToRevision(payload, validChangeKeys) {
    if (!isRawRevisionPayload(payload)) {
        throw new Error('Invalid revision payload');
    }
    let filterChanges = (cc) => cc;
    if (validChangeKeys.length > 0) {
        // filter out changes that don't have valid keys
        filterChanges = (cc) => cc.filter(c => validChangeKeys.includes(c.key));
    }
    return payload.set.map(raw => ({
        changeID: raw.changeID,
        timestamp: new Date(raw.timestamp),
        resource: raw.resource,
        revision: raw.revision,
        operation: raw.operation,
        userID: raw.userID,
        user: null,
        comment: raw.comment,
        changes: filterChanges(raw.changes),
    }));
}

function generateUID() {
    const uid = Math.random().toString(36).substring(2) + (new Date()).getTime().toString(36);
    return `tempID-${uid}`;
}

const { merge: merge$7 } = lodash;
const defaultXYWH = [0, 0, 20, 15];
class PageBlock {
    constructor(i) {
        // blockID is auto generated by the server in order to support resource translations
        this.blockID = NoID;
        this.kind = '';
        this.title = '';
        this.description = '';
        this.xywh = defaultXYWH;
        this.options = {};
        this.meta = {
            hidden: false,
            tempID: undefined,
            customID: undefined,
            customCSSClass: undefined,
            visibility: {
                expression: '',
                roles: [],
            },
        };
        this.style = {
            variants: {
                headerText: 'dark',
            },
            wrap: {
                kind: 'card',
            },
            border: {
                enabled: false,
            },
        };
        this.apply(i);
        this.setTempID();
    }
    apply(i) {
        if (!i)
            return;
        Apply(this, i, String, 'title', 'description', 'blockID');
        if (i.xywh) {
            if (!Array.isArray(i.xywh)) {
                throw new Error('xywh must be an array');
            }
            if (i.xywh.length !== 4) {
                throw new Error('xywh must have 4 elements');
            }
            // by default, park 3x3 block in upper left corner
            this.xywh = i.xywh || defaultXYWH;
        }
        if (i.options) {
            this.options = merge$7({}, this.options, i.options);
        }
        if (i.style) {
            this.style = merge$7({}, this.style, i.style);
        }
        if (i.meta) {
            this.meta = merge$7({}, this.meta, i.meta);
        }
    }
    // Returns Page Block configuration errors
    validate() {
        return [];
    }
    setTempID() {
        this.meta.tempID = this.meta.tempID || generateUID();
    }
    clone() {
        return new this.constructor(Object.assign(Object.assign({}, JSON.parse(JSON.stringify(this))), { blockID: NoID, meta: Object.assign(Object.assign({}, this.meta), { tempID: '' }) }));
    }
}
const Registry = new Map();

class Button {
    constructor(b) {
        // Used when referring to Corredor automation script
        this.script = undefined;
        // Used when referring to workflow with onManual trigger
        this.workflowID = undefined;
        // Used when referring to a specific step (triggered by onManual trigger)
        this.stepID = undefined;
        // resource type (copied from ui hook or from trigger)
        this.resourceType = undefined;
        // Can override hook's label
        this.label = undefined;
        // can override hook's variant
        this.variant = 'primary';
        this.enabled = true;
        Apply(this, b, Boolean, 'enabled');
        Apply(this, b, String, 'label', 'variant', 'script', 'resourceType');
        Apply(this, b, CortezaID, 'workflowID', 'stepID');
    }
}

const kind$h = 'Automation';
const defaults$h = Object.freeze({
    buttons: [],
    sealed: false,
    magnifyOption: '',
});
class PageBlockAutomation extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$h;
        this.options = Object.assign({}, defaults$h);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, String, 'magnifyOption');
        if (o.buttons) {
            this.options.buttons = o.buttons.map(b => new Button(b));
        }
    }
    // Validates Page Block configuration
    validate() {
        const ee = super.validate();
        this.options.buttons.forEach(b => {
            if (b.workflowID) {
                // workflow defined
                return;
            }
            if (b.script) {
                // script defined
                return;
            }
            ee.push('Automation button without configured script or workflow');
        });
        return ee;
    }
}
Registry.set(kind$h, PageBlockAutomation);

const { cloneDeep, merge: merge$6 } = lodash;
const kind$g = 'Chart';
const defaults$g = Object.freeze({
    chartID: '',
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
    liveFilterEnabled: false,
    drillDown: {
        enabled: false,
        blockID: '',
        recordListOptions: {
            fields: [],
        },
    },
});
class PageBlockChart extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$g;
        this.options = Object.assign({}, defaults$g);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        o.chartID = o.chartID === NoID ? '' : o.chartID;
        Apply(this.options, o, String, 'chartID', 'magnifyOption');
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, Boolean, 'showRefresh', 'liveFilterEnabled');
        if (o.drillDown) {
            this.options.drillDown = merge$6({}, defaults$g.drillDown, o.drillDown);
        }
    }
    resetDrillDown() {
        this.options.drillDown = cloneDeep(defaults$g.drillDown);
    }
}
Registry.set(kind$g, PageBlockChart);

const kind$f = 'Content';
const defaults$f = Object.freeze({
    body: '',
    magnifyOption: '',
});
class PageBlockContent extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$f;
        this.options = Object.assign({}, defaults$f);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, String, 'body', 'magnifyOption');
    }
}
Registry.set(kind$f, PageBlockContent);

const kind$e = 'File';
const PageBlockFileDefaultMode = 'list';
const PageBlockFileModes = [
    // list of attachments, no preview
    'list',
    // list of all images/files, show preview
    'gallery',
];
const defaults$e = Object.freeze({
    mode: PageBlockFileDefaultMode,
    attachments: [],
    hideFileName: false,
    height: '',
    width: '',
    maxHeight: '',
    maxWidth: '',
    borderRadius: '',
    margin: 'auto',
    backgroundColor: '#FFFFFF00',
    magnifyOption: '',
    clickToView: true,
    enableDownload: true,
});
class PageBlockFile extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$e;
        this.options = Object.assign({}, defaults$e);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        if (o.attachments) {
            this.options.attachments = o.attachments;
        }
        Apply(this.options, o, Boolean, 'hideFileName', 'clickToView', 'enableDownload');
        Apply(this.options, o, String, 'height', 'width', 'maxHeight', 'maxWidth', 'borderRadius', 'margin', 'backgroundColor', 'magnifyOption');
        if (o.mode) {
            // Legacy
            if (o.mode === 'single') {
                o.mode = 'gallery';
            }
            else if (o.mode === 'grid') {
                o.mode = 'list';
            }
            if (PageBlockFileModes.includes(o.mode)) {
                this.options.mode = o.mode;
            }
            else {
                o.mode = PageBlockFileDefaultMode;
            }
        }
    }
}
Registry.set(kind$e, PageBlockFile);

const kind$d = 'IFrame';
const defaults$d = Object.freeze({
    srcField: '',
    src: '',
    wrap: 'Plain',
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
});
class PageBlockIFrame extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$d;
        this.options = Object.assign({}, defaults$d);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, String, 'srcField', 'src', 'wrap', 'magnifyOption');
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, Boolean, 'showRefresh');
    }
}
Registry.set(kind$d, PageBlockIFrame);

const kind$c = 'Record';
const defaults$c = Object.freeze({
    fields: [],
    fieldConditions: [],
    clearConditionalFieldsOnHide: false,
    recordSelectorShowAddRecordButton: false,
    magnifyOption: '',
    recordSelectorDisplayOption: 'sameTab',
    recordSelectorAddRecordDisplayOption: 'sameTab',
    referenceField: '',
    referenceModuleID: undefined,
    inlineRecordEditEnabled: false,
    inlineRecordEditAllowAddField: false,
    horizontalFieldLayoutEnabled: false,
    recordFieldLayoutOption: 'default',
});
class PageBlockRecord extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$c;
        this.options = Object.assign({}, defaults$c);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, String, 'magnifyOption', 'recordSelectorDisplayOption', 'recordSelectorAddRecordDisplayOption', 'referenceField', 'referenceModuleID', 'recordFieldLayoutOption');
        Apply(this.options, o, Boolean, 'recordSelectorShowAddRecordButton', 'inlineRecordEditEnabled', 'horizontalFieldLayoutEnabled', 'inlineRecordEditAllowAddField', 'clearConditionalFieldsOnHide');
        if (o.fields) {
            this.options.fields = o.fields;
        }
        if (o.fieldConditions) {
            this.options.fieldConditions = o.fieldConditions;
        }
    }
}
Registry.set(kind$c, PageBlockRecord);

const kind$b = 'RecordList';
var SummaryMetric;
(function (SummaryMetric) {
    SummaryMetric["Min"] = "min";
    SummaryMetric["Max"] = "max";
    SummaryMetric["Avg"] = "avg";
    SummaryMetric["Sum"] = "sum";
    SummaryMetric["EmptyCount"] = "emptyCount";
    SummaryMetric["NotEmptyCount"] = "notEmptyCount";
    SummaryMetric["UniqueCount"] = "uniqueCount";
    SummaryMetric["Earliest"] = "earliest";
    SummaryMetric["Latest"] = "latest";
})(SummaryMetric || (SummaryMetric = {}));
const defaults$b = Object.freeze({
    moduleID: NoID,
    prefilter: '',
    presort: 'createdAt DESC',
    fields: [],
    inlineEditFields: [],
    hideHeader: false,
    hideAddButton: false,
    hideImportButton: false,
    hideConfigureFieldsButton: true,
    hideSearch: false,
    hidePaging: false,
    hideSorting: false,
    hideFiltering: false,
    hideRecordReminderButton: false,
    hideRecordCloneButton: false,
    hideRecordEditButton: false,
    hideRecordViewButton: false,
    hideRecordPermissionsButton: false,
    hideRecordDeleteButton: false,
    enableRecordPageNavigation: true,
    allowExport: true,
    perPage: 20,
    recordDisplayOption: 'sameTab',
    recordSelectorDisplayOption: 'sameTab',
    addRecordDisplayOption: 'sameTab',
    magnifyOption: '',
    searchableFields: [],
    fullPageNavigation: false,
    showTotalCount: true,
    showDeletedRecordsOption: false,
    customFilterPresets: false,
    editable: false,
    draggable: false,
    positionField: undefined,
    refField: undefined,
    editFields: [],
    linkToParent: false,
    openInNewTab: false,
    selectable: true,
    selectMode: 'multi',
    selectionButtons: [],
    refreshRate: 0,
    showRefresh: false,
    bulkRecordEditEnabled: true,
    inlineRecordEditEnabled: false,
    inlineRecordEditAllowAddField: false,
    inlineValueFiltering: false,
    filterPresets: [],
    showRecordPerPageOption: false,
    openRecordInEditMode: false,
    customSummaries: false,
    summaries: [],
    textStyles: {
        wrappedFields: [],
    },
});
class PageBlockRecordList extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$b;
        this.options = Object.assign({}, defaults$b);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, CortezaID, 'moduleID');
        Apply(this.options, o, String, 'prefilter', 'presort', 'selectMode', 'positionField', 'refField', 'recordDisplayOption', 'magnifyOption', 'recordSelectorDisplayOption', 'addRecordDisplayOption');
        Apply(this.options, o, Number, 'perPage', 'refreshRate');
        if (o.fields) {
            this.options.fields = o.fields;
        }
        if (o.searchableFields) {
            this.options.searchableFields = o.searchableFields;
        }
        if (o.inlineEditFields) {
            this.options.inlineEditFields = o.inlineEditFields;
        }
        if (o.filterPresets) {
            this.options.filterPresets = o.filterPresets;
        }
        if (o.editFields) {
            this.options.editFields = o.editFields;
        }
        if (o.openInNewTab) {
            this.options.recordDisplayOption = 'newTab';
        }
        Apply(this.options, o, Boolean, 'hideHeader', 'hideAddButton', 'hideImportButton', 'hideConfigureFieldsButton', 'hideSearch', 'hidePaging', 'hideFiltering', 'fullPageNavigation', 'showTotalCount', 'showDeletedRecordsOption', 'customFilterPresets', 'hideSorting', 'allowExport', 'selectable', 'hideRecordReminderButton', 'hideRecordCloneButton', 'hideRecordEditButton', 'hideRecordViewButton', 'hideRecordPermissionsButton', 'hideRecordDeleteButton', 'enableRecordPageNavigation', 'editable', 'draggable', 'linkToParent', 'showRefresh', 'bulkRecordEditEnabled', 'inlineRecordEditEnabled', 'inlineRecordEditAllowAddField', 'inlineValueFiltering', 'showRecordPerPageOption', 'openRecordInEditMode', 'customSummaries');
        if (o.selectionButtons) {
            this.options.selectionButtons = o.selectionButtons.map(b => new Button(b));
        }
        if (o.summaries) {
            this.options.summaries = o.summaries;
        }
        if (o.textStyles) {
            this.options.textStyles = Object.assign(Object.assign({}, this.options.textStyles), o.textStyles);
        }
    }
    fetch(api, recordListModule, filter) {
        return __awaiter(this, void 0, void 0, function* () {
            if (recordListModule.moduleID !== this.options.moduleID) {
                throw Error('Module incompatible, module mismatch');
            }
            filter.moduleID = this.options.moduleID;
            filter.namespaceID = recordListModule.namespaceID;
            return api
                .recordList(filter)
                .then(r => {
                const { set: records, filter } = r;
                return { records, filter };
            });
        });
    }
}
Registry.set(kind$b, PageBlockRecordList);

const kind$a = 'RecordRevisions';
const defaults$a = Object.freeze({
    preload: false,
    displayedFields: [],
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
    sortDirection: 'desc',
});
class PageBlockRecordRevisions extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$a;
        this.options = Object.assign({}, defaults$a);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, Boolean, 'preload', 'showRefresh');
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, String, 'magnifyOption', 'sortDirection');
        // set new values to displayed fields
        if (Array.isArray(o === null || o === void 0 ? void 0 : o.displayedFields)) {
            this.options.displayedFields = o.displayedFields.map(String);
        }
    }
    /**
     * fetch is a utility method on record revision page block
     * that fetches revisions for a record and converts them to RevisionPayload class
     *
     * this function also strips out all fields that should not be dispalyed
     * (as per displayedFields option)
     *
     * @param api Compose API to be used
     * @param record Record to fetch revisions for
     * @param sortDirection Sort direction ('asc' for oldest first, 'desc' for newest first)
     */
    fetch(api, record, sortDirection) {
        return __awaiter(this, void 0, void 0, function* () {
            const { namespaceID, moduleID, recordID } = record;
            // Build sort parameter based on sortDirection
            // Default to 'desc' (newest first) if not specified
            const sort = sortDirection === 'asc' ? 'revision ASC' : 'revision DESC';
            return api
                .recordRevisions({ namespaceID, moduleID, recordID, sort })
                .then(payload => convertRevisionPayloadToRevision(payload, this.options.displayedFields));
        });
    }
}
Registry.set(kind$a, PageBlockRecordRevisions);

const kind$9 = 'RecordOrganizer';
const defaults$9 = Object.freeze({
    moduleID: NoID,
    labelField: '',
    descriptionField: '',
    filter: '',
    positionField: '',
    groupField: '',
    group: '',
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
    displayOption: 'sameTab',
});
class PageBlockRecordOrganizer extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$9;
        this.options = Object.assign({}, defaults$9);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, CortezaID, 'moduleID');
        Apply(this.options, o, String, 'labelField', 'descriptionField', 'filter', 'positionField', 'groupField', 'group', 'magnifyOption', 'displayOption');
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, Boolean, 'showRefresh');
    }
}
Registry.set(kind$9, PageBlockRecordOrganizer);

const kind$8 = 'SocialFeed';
const defaults$8 = Object.freeze({
    moduleID: NoID,
    fields: [],
    profileSourceField: '',
    profileUrl: '',
    showRefresh: false,
    refreshRate: 0,
    magnifyOption: '',
});
class PageBlockSocialFeed extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$8;
        this.options = Object.assign({}, defaults$8);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, CortezaID, 'moduleID');
        Apply(this.options, o, String, 'profileSourceField', 'profileUrl', 'magnifyOption');
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, Boolean, 'showRefresh');
        if (o.fields) {
            this.options.fields = o.fields;
        }
    }
}
Registry.set(kind$8, PageBlockSocialFeed);

const feedResources = {
    record: 'compose:record'};

const defOptions$2 = {
    moduleID: NoID,
    color: '#61AFFF',
    prefilter: '',
};
/**
 * Feed class represents an event feed for the given calendar
 */
let Feed$1 = class Feed {
    constructor(i) {
        this.resource = 'compose:record';
        this.startField = '';
        this.endField = '';
        this.titleField = '';
        this.options = Object.assign({}, defOptions$2);
        this.allDay = false;
        this.apply(i);
    }
    apply(i) {
        if (!i)
            return;
        if (!IsOf(i, 'resource') && IsOf(i, 'moduleID')) {
            i = Feed.fromLegacy(i);
        }
        if (IsOf(i, 'resource')) {
            Apply(this, i, String, 'resource', 'startField', 'endField', 'titleField');
            Apply(this, i, Boolean, 'allDay');
            if (i.options) {
                this.options = Object.assign(Object.assign({}, this.options), i.options);
            }
        }
    }
    static fromLegacy(legacy) {
        const p = Object.assign({ 
            // legacy does not have resource,
            // we've used it with records only
            resource: feedResources.record }, legacy);
        if (legacy.moduleID) {
            if (!p.options) {
                p.options = Object.assign({}, defOptions$2);
            }
            // module was moved under options
            p.options.moduleID = legacy.moduleID;
        }
        return p;
    }
};

const isLightThreshold = 100;
/**
 * Helper to determine event's colors
 * @param {String} hex Base color in HEX format
 * @returns {Object} { backgroundColor: String, borderColor: String, isLight: Boolean }
 */
function makeColors(hex) {
    const bg = hr(hex, { format: 'array' });
    const br = [...bg];
    const isLight = (bg.slice(0, 3).reduce((acc, cur) => acc + cur, 0) / (bg.length - 1)) > isLightThreshold;
    return {
        textColor: isLight ? '#000' : '#fff',
        backgroundColor: `rgba(${bg.join(',')})`,
        borderColor: `rgba(${br.join(',')})`,
    };
}

/**
 * Loads & converts reminder resource into FC events
 * @param {SystemAPI} $SystemAPI SystemAPI provider
 * @param {User} user Current user
 * @param {Feed} feed Current feed
 * @param {Object} range Current date range
 * @returns {Promise<Array>} Resolves to a set of FC events to display
 */
function ReminderFeed($SystemAPI_1, user_1, feed_1, range_1) {
    return __awaiter(this, arguments, void 0, function* ($SystemAPI, user, feed, range, options = {}) {
        return $SystemAPI.reminderList({
            scheduledFrom: range.start.toISOString(),
            scheduledUntil: range.end.toISOString(),
            scheduledOnly: true,
            excludeDismissed: true,
            assignedTo: user.userID,
        }, options).then(({ set }) => {
            const { backgroundColor, borderColor, textColor } = makeColors(feed.options.color);
            if (!AreObjectsOf(set, 'reminderID', 'assignedTo', 'remindAt', 'payload')) {
                return [];
            }
            return set.map(r => {
                var _a, _b, _c;
                r = new Reminder(r);
                const classNames = ['event', 'event-reminder'];
                if (r.assignedTo !== user.userID) {
                    classNames.push('event-not-owner');
                }
                const e = {
                    id: r.reminderID,
                    title: (_b = (_a = r.payload.title) !== null && _a !== void 0 ? _a : r.reminderID) !== null && _b !== void 0 ? _b : '-',
                    start: (_c = r.remindAt) === null || _c === void 0 ? void 0 : _c.toISOString(),
                    backgroundColor,
                    borderColor,
                    textColor,
                    classNames,
                    allDay: false,
                    extendedProps: {
                        reminderID: r.reminderID,
                    },
                };
                return e;
            });
        });
    });
}

function getRecordValue(record, field) {
    const ef = record.module.fields.find(({ name }) => name === field);
    if (ef) {
        return ef.isMulti ? record.values[field] : [record.values[field] || undefined];
    }
    else {
        switch (field) {
            case 'recordID':
            case 'moduleID':
            case 'namespaceID':
                return [record[field]];
            case 'createdAt':
            case 'updatedAt':
            case 'deletedAt':
                if (record[field] !== undefined) {
                    return [record[field].toISOString()];
                }
                break;
        }
    }
    return [undefined];
}
/**
 * Method expands the given record in a (set) of FC event objects.
 * Handles basic recurrence -- multiple date fields.
 * @param {Record} record Record to expand
 * @param {Feed} feed Feed, this record belongs to
 * @returns {Array} A set of expanded events
 */
function expandRecord(record, feed) {
    const events = [];
    const starts = getRecordValue(record, feed.startField);
    const ends = getRecordValue(record, feed.endField);
    const title = getRecordValue(record, feed.titleField).shift() || record.recordID;
    // Make sure ends is at least as long as starts, to avoid length checks
    ends.push(...(new Array(Math.max(starts.length - ends.length, 0)).fill(undefined)));
    const classNames = ['event', 'event-record'];
    const { backgroundColor, borderColor, textColor } = makeColors(feed.options.color);
    starts.forEach((start, i) => {
        events.push({
            // So FC knows how to group these expanded events
            groupId: record.recordID,
            id: record.recordID,
            title,
            start: start,
            end: ends[i],
            allDay: feed.allDay,
            backgroundColor,
            borderColor,
            classNames,
            textColor,
            extendedProps: {
                moduleID: record.module.moduleID,
                recordID: record.recordID,
            },
        });
    });
    return events;
}
/**
 * Checks if the given field can be used with the given record.
 * A field can be used if it's either defined as a record value OR it's a system field.
 *
 * @param r The record to check
 * @param field The field we wish to use
 */
function recordFeedFilter(r, field) {
    if (r.values[field]) {
        return true;
    }
    return ['createdAt', 'updatedAt', 'deletedAt'].includes(field);
}
/**
 * Loads & converts module resource into FC events
 * @param {ComposeAPI} $ComposeAPI ComposeAPI provider
 * @param {Module} module Current module
 * @param {Namespace} namespace Current namespace
 * @param {Feed} feed Current feed
 * @param {Object} range Current date range
 * @returns {Promise<Array>} Resolves to a set of FC events to display
 */
function RecordFeed$1($ComposeAPI_1, module_1, namespace_1, feed_1, range_1) {
    return __awaiter(this, arguments, void 0, function* ($ComposeAPI, module, namespace, feed, range, options = {}) {
        // Params for record fetching
        const params = {
            namespaceID: namespace.namespaceID,
            moduleID: module.moduleID,
            query: `(date(${feed.startField}) <= '${range.end.toISOString()}' AND '${range.start.toISOString()}' <= date(${feed.endField || feed.startField}))`,
        };
        if (feed.options.prefilter) {
            params.query += ` AND (${feed.options.prefilter})`;
        }
        const events = [];
        return $ComposeAPI.recordList(params, options).then(({ set }) => {
            set
                // Removes all duplicates
                .filter(({ recordID }, index, set) => set.findIndex((r) => recordID === r.recordID) === index)
                // cast & freeze
                .map(r => Object.freeze(new Record(module, r)))
                // drop record w/o proper values
                .filter(r => recordFeedFilter(r, feed.startField))
                // eslint-disable-next-line @typescript-eslint/no-use-before-define
                .forEach(r => events.push(...expandRecord(r, feed)));
            return events;
        });
    });
}

const { merge: merge$5 } = lodash;
const kind$7 = 'Calendar';
// Map of < V4 view names to >= V4 view names
const legacyViewMapping = {
    month: 'dayGridMonth',
    agendaMonth: 'dayGridMonth',
    agendaWeek: 'timeGridWeek',
    agendaDay: 'timeGridDay',
    listMonth: 'listMonth',
};
const defaults$7 = Object.freeze({
    defaultView: '',
    feeds: [],
    header: {},
    locale: 'en-gb',
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
    eventDisplayOption: 'sameTab',
});
/**
 * Helper class to help define calendar's functionality
 */
class PageBlockCalendar extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$7;
        this.options = Object.assign({}, defaults$7);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        var _a;
        if (!o)
            return;
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, Boolean, 'showRefresh');
        Apply(this.options, o, String, 'magnifyOption', 'eventDisplayOption');
        this.options.defaultView = PageBlockCalendar.handleLegacyView(o.defaultView) || 'dayGridMonth';
        this.options.feeds = (o.feeds || []).map(f => new Feed$1(f));
        this.options.header = merge$5({}, this.options.header, o.header, { views: PageBlockCalendar.handleLegacyViews(((_a = o.header) === null || _a === void 0 ? void 0 : _a.views) || ['dayGridMonth', 'timeGridWeek', 'timeGridDay', 'listMonth']) });
        this.options.locale = o.locale || 'en-gb';
    }
    /**
     * Generates a header object of fullcalendar
     * @returns {Object}
     */
    getHeader() {
        const h = this.options.header;
        if (h.hide) {
            return;
        }
        // Show view buttons only when 2 or more are selected
        let right = '';
        if (h.views && h.views.length >= 2) {
            right = this.reorderViews(h.views).join(',');
        }
        return {
            left: `${h.hidePrevNext ? '' : 'prev,next'} ${h.hideToday ? '' : 'today'}`.trim(),
            center: `${h.hideTitle ? '' : 'title'}`,
            right,
        };
    }
    /**
     * Provides a list of available views.
     * @note When adding new ones, make sure included plugins support it.
     * @returns {Array}
     */
    static availableViews() {
        return [
            'dayGridMonth',
            'timeGridWeek',
            'timeGridDay',
            'listMonth',
        ];
    }
    /**
     * Reorder views according to available views array order.
     * @param {Array} views Array of views to filter & sort
     */
    reorderViews(views = []) {
        return PageBlockCalendar.availableViews()
            .filter(v => views.find(fv => fv === v))
            .map(v => v);
    }
    /**
     * Converts old < V4 view names to >= V4 view names.
     * @note It wil preserve fields that don't need to/can't be converted
     * @param {string} views converted view name
     */
    static handleLegacyView(views = 'dayGridMonth') {
        return legacyViewMapping[views] || views;
    }
    /**
     * Converts old < V4 view names to >= V4 view names.
     * @note It wil preserve fields that don't need to/can't be converted
     * @param {string[]} views converted view names
     */
    static handleLegacyViews(views) {
        return views.map(v => legacyViewMapping[v] || v);
    }
    static makeFeed(f) {
        return new Feed$1(f);
    }
}
PageBlockCalendar.feedResources = Object.freeze({
    record: 'compose:record',
    reminder: 'system:reminder',
});
PageBlockCalendar.ReminderFeed = ReminderFeed;
PageBlockCalendar.RecordFeed = RecordFeed$1;
Registry.set(kind$7, PageBlockCalendar);

const { merge: merge$4 } = lodash;
const kind$6 = 'Metric';
const defaultMetric = Object.freeze({
    label: '',
    moduleID: '',
    dimensionField: '',
    dateFormat: '',
    filter: '',
    bucketSize: '',
    metricField: '',
    operation: '',
    numberFormat: '',
    prefix: '',
    suffix: '',
    transformFx: '',
    valueStyle: {
        backgroundColor: '#FFFFFF00',
        color: '#000000',
        fontSize: undefined,
    },
    drillDown: {
        enabled: false,
        blockID: '',
        recordListOptions: {
            fields: [],
        },
    },
});
const defaults$6 = Object.freeze({
    metrics: [],
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
});
class PageBlockMetric extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$6;
        this.options = Object.assign({}, defaults$6);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, Boolean, 'showRefresh');
        Apply(this.options, o, String, 'magnifyOption');
        if (o.metrics) {
            this.options.metrics = o.metrics.map((m) => merge$4({}, defaultMetric, m));
        }
    }
    /**
     * Helper function to fetch and parse reporter's reports.
     */
    fetch(_a, reporter_1) {
        return __awaiter(this, arguments, void 0, function* ({ m }, reporter) {
            const w = yield reporter(this.formatParams(m));
            const datasets = w.map((r) => r.rp !== undefined ? r.rp : r.count);
            let rtr;
            if (m.operation === 'max') {
                rtr = datasets.sort((a, b) => b - a)[0];
            }
            else if (m.operation === 'min') {
                rtr = datasets.sort((a, b) => a - b)[0];
            }
            else if (m.operation === 'avg') {
                rtr = datasets.reduce((acc, cur) => acc + cur, 0) / datasets.length;
            }
            else {
                rtr = datasets.reduce((acc, cur) => acc + cur, 0);
            }
            if (m.transformFx) {
                // eslint-disable-next-line no-new-func
                rtr = (new Function('v', `return ${m.transformFx}`))(rtr);
            }
            return [{ value: rtr }];
        });
    }
    /**
     * Helper to construct reporter's params
     */
    formatParams({ moduleID, filter, metricField, operation = '' }) {
        let metrics = '';
        if (operation && metricField && metricField !== 'count') {
            metrics = `${operation}(${metricField}) AS rp`;
        }
        return {
            moduleID,
            filter,
            metrics,
            // Since metric produces one value we want one dataset, deletedAt is the same for all existing records
            dimensions: 'deletedAt',
        };
    }
    makeMetric() {
        return merge$4({}, defaultMetric);
    }
}
Registry.set(kind$6, PageBlockMetric);

const kind$5 = 'Comment';
const defaults$5 = Object.freeze({
    moduleID: NoID,
    filter: '',
    titleField: '',
    contentField: '',
    replyField: '',
    sortDirection: 'asc',
    referenceField: '',
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
    attachmentField: '',
    reactionsField: '',
});
class PageBlockComment extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$5;
        this.options = Object.assign({}, defaults$5);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, CortezaID, 'moduleID');
        Apply(this.options, o, String, 'titleField', 'contentField', 'replyField', 'referenceField', 'attachmentField', 'reactionsField', 'filter', 'sortDirection', 'magnifyOption');
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, Boolean, 'showRefresh');
    }
}
Registry.set(kind$5, PageBlockComment);

const kind$4 = 'Report';
const defaults$4 = Object.freeze({
    reportID: NoID,
    scenarioID: NoID,
    elementID: NoID,
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
});
class PageBlockReport extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$4;
        this.options = Object.assign({}, defaults$4);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, CortezaID, 'reportID', 'scenarioID', 'elementID');
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, Boolean, 'showRefresh');
        Apply(this.options, o, String, 'magnifyOption');
    }
}
Registry.set(kind$4, PageBlockReport);

var ChartType;
(function (ChartType) {
    ChartType["pie"] = "pie";
    ChartType["bar"] = "bar";
    ChartType["line"] = "line";
    ChartType["doughnut"] = "doughnut";
    ChartType["funnel"] = "funnel";
    ChartType["gauge"] = "gauge";
    ChartType["radar"] = "radar";
    ChartType["scatter"] = "scatter";
})(ChartType || (ChartType = {}));
const aggregateFunctions = [
    {
        value: 'SUM',
        text: 'sum',
    },
    {
        value: 'MAX',
        text: 'max',
    },
    {
        value: 'MIN',
        text: 'min',
    },
    {
        value: 'AVG',
        text: 'avg',
    },
    {
        value: 'STD',
        text: 'std',
    },
];
class DimensionFunctions extends Array {
    constructor(items) {
        super(...(items || []));
    }
    static create() {
        return Object.create(DimensionFunctions.prototype);
    }
    lookup(d) {
        return this.find((f) => d.modifier === f.value);
    }
    convert(d) {
        return (this.lookup(d) || {}).convert(d.field);
    }
}
const dimensionFunctions = DimensionFunctions.create();
dimensionFunctions.push(...[
    {
        text: 'none',
        value: '(no grouping / buckets)',
        convert: (f) => f,
    },
    {
        text: 'date',
        value: 'DATE',
        convert: (f) => `DATE(${f})`,
    },
    {
        text: 'week',
        value: 'WEEK',
        convert: (f) => `WEEK(${f})`,
    },
    {
        text: 'month',
        value: 'MONTH',
        convert: (f) => `DATE_FORMAT(${f}, '%Y-%m-01')`,
    },
    {
        text: 'quarter',
        value: 'QUARTER',
        convert: (f) => `QUARTER(${f})`,
    },
    {
        text: 'year',
        value: 'YEAR',
        convert: (f) => `DATE_FORMAT(${f}, '%Y-01-01')`,
    },
]);
const predefinedFilters = [
    {
        value: 'YEAR(createdAt) = YEAR(NOW())',
        text: 'recordsCreatedThisYear',
    },
    {
        value: 'YEAR(createdAt) = YEAR(NOW()) - 1',
        text: 'recordsCreatedLastYear',
    },
    {
        value: 'YEAR(createdAt) = YEAR(NOW()) AND QUARTER(createdAt) = QUARTER(NOW())',
        text: 'recordsCreatedThisQuarter',
    },
    {
        value: 'YEAR(createdAt) = YEAR(NOW()) AND QUARTER(createdAt) = QUARTER(DATE_SUB(NOW(), INTERVAL 3 MONTH)',
        text: 'recordsCreatedLastQuarter',
    },
    {
        value: 'DATE_FORMAT(createdAt, \'%Y-%m\') = DATE_FORMAT(NOW(), \'%Y-%m\')',
        text: 'recordsCreatedThisMonth',
    },
    {
        value: 'DATE_FORMAT(createdAt, \'%Y-%m\') = DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 MONTH), \'%Y-%m\')',
        text: 'recordsCreatedLastMonth',
    },
];
dimensionFunctions.lookup = d => dimensionFunctions.find(f => d.modifier === f.value) || dimensionFunctions[0];
dimensionFunctions.convert = d => dimensionFunctions.lookup(d).convert(d.field);
const isRadialChart = ({ type }) => type === 'doughnut' || type === 'pie';
const hasRelativeDisplay = ({ type }) => isRadialChart({ type });
// Makes a standardized alias from modifier or dimension report option
const makeAlias = ({ alias, aggregate, modifier, field }) => alias || `${aggregate || modifier || 'none'}_${field}`.toLocaleLowerCase();
function formatChartValue(value, formatting) {
    var _a, _b;
    let n = '';
    // if value contains alphabetic chars parseFloat() will return NaN
    // and n will equal 0
    const containsAlphabeticChars = isNaN(Number(value));
    let result = '';
    if (!containsAlphabeticChars) {
        switch (typeof value) {
            case 'string':
                n = parseFloat(value);
                break;
            case 'number':
                n = value;
                break;
            default:
                n = 0;
        }
        if (formatting === null || formatting === void 0 ? void 0 : formatting.format) {
            result = numeral(n).format(formatting.format);
        }
        else {
            result = number(n);
        }
    }
    if ((formatting === null || formatting === void 0 ? void 0 : formatting.presetFormat) === 'accounting') {
        result = accountingNumber(Number(n));
    }
    return ` ${(_a = formatting === null || formatting === void 0 ? void 0 : formatting.prefix) !== null && _a !== void 0 ? _a : ''} ${result || value} ${(_b = formatting === null || formatting === void 0 ? void 0 : formatting.suffix) !== null && _b !== void 0 ? _b : ''}`;
}
function formatChartTooltip(tooltip, params) {
    const { seriesName = '', name = '', value = '', percent = '' } = params;
    return tooltip
        .replace('{a}', seriesName)
        .replace('{b}', name)
        .replace('{c}', value.toString())
        .replace('{d}', percent.toString());
}
function defFormatData() {
    return Object.assign({}, {
        presetFormat: 'custom',
        prefix: '',
        suffix: '',
        format: '',
    });
}
const chartUtil = {
    dimensionFunctions,
    hasRelativeDisplay,
    aggregateFunctions,
    predefinedFilters,
    ChartType,
};

const kind$3 = 'Progress';
const defaults$3 = Object.freeze({
    value: {
        default: 0,
        moduleID: '',
        filter: '',
        field: '',
        operation: '',
    },
    minValue: {
        default: 0,
        moduleID: '',
        filter: '',
        field: '',
        operation: '',
    },
    maxValue: {
        default: 100,
        moduleID: '',
        filter: '',
        field: '',
        operation: '',
    },
    display: {
        showValue: true,
        showRelative: true,
        showProgress: false,
        animated: false,
        variant: 'success',
        thresholds: [],
    },
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
});
class PageBlockProgress extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$3;
        this.options = Object.assign({}, defaults$3);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, Number, 'refreshRate');
        Apply(this.options, o, Boolean, 'showRefresh');
        Apply(this.options, o, String, 'magnifyOption');
        if (o.value) {
            this.options.value = Object.assign(Object.assign({}, this.options.value), o.value);
        }
        if (o.minValue) {
            this.options.minValue = Object.assign(Object.assign({}, this.options.minValue), o.minValue);
        }
        if (o.maxValue) {
            this.options.maxValue = Object.assign(Object.assign({}, this.options.maxValue), o.maxValue);
        }
        if (o.display) {
            this.options.display = Object.assign(Object.assign({}, this.options.display), o.display);
        }
    }
    /**
     * Helper function to fetch and parse reporter's reports.
     */
    fetch(additionalOptions, api, namespaceID) {
        const reports = [];
        const dimensions = dimensionFunctions.convert({ modifier: 'YEAR', field: 'createdAt' });
        let metrics = '';
        // Construct value report
        const { field: valueField, operation: valueOperation = '' } = this.options.value;
        if (this.options.value.moduleID && valueField) {
            if (valueOperation && valueField !== 'count') {
                metrics = `${valueOperation}(${valueField}) AS rp`;
            }
            reports.push(api.recordReport(Object.assign(Object.assign({ namespaceID, metrics, dimensions }, this.options.value), additionalOptions.value)));
        }
        else {
            reports.push(new Promise(resolve => resolve(this.options.value.default)));
        }
        // Construct minValue report
        const { field: minValueField, operation: minValueOperation = '' } = this.options.minValue;
        if (this.options.minValue.moduleID && minValueField) {
            metrics = '';
            if (minValueOperation && minValueField !== 'count') {
                metrics = `${minValueOperation}(${minValueField}) AS rp`;
            }
            reports.push(api.recordReport(Object.assign(Object.assign({ namespaceID, metrics, dimensions }, this.options.minValue), additionalOptions.minValue)));
        }
        else {
            reports.push(new Promise(resolve => resolve(this.options.minValue.default)));
        }
        // Construct minValue report
        const { field: maxValueField, operation: maxValueOperation = '' } = this.options.maxValue;
        if (this.options.maxValue.moduleID && maxValueField) {
            metrics = '';
            if (maxValueOperation && maxValueField !== 'count') {
                metrics = `${maxValueOperation}(${maxValueField}) AS rp`;
            }
            reports.push(api.recordReport(Object.assign(Object.assign({ namespaceID, metrics, dimensions }, this.options.maxValue), additionalOptions.maxValue)));
        }
        else {
            reports.push(new Promise(resolve => resolve(this.options.maxValue.default)));
        }
        return Promise.all(reports).then(([value, min, max]) => {
            if (Array.isArray(value)) {
                const datasets = value.map((r) => r.rp !== undefined ? r.rp : r.count);
                if (valueOperation === 'max') {
                    value = datasets.sort((a, b) => b - a)[0];
                }
                else if (valueOperation === 'min') {
                    value = datasets.sort((a, b) => a - b)[0];
                }
                else if (valueOperation === 'avg') {
                    value = datasets.reduce((acc, cur) => acc + cur, 0) / datasets.length;
                }
                else {
                    value = datasets.reduce((acc, cur) => acc + cur, 0);
                }
            }
            if (Array.isArray(min)) {
                const datasets = min.map((r) => r.rp !== undefined ? r.rp : r.count);
                if (minValueOperation === 'max') {
                    min = datasets.sort((a, b) => b - a)[0];
                }
                else if (minValueOperation === 'min') {
                    min = datasets.sort((a, b) => a - b)[0];
                }
                else if (minValueOperation === 'avg') {
                    min = datasets.reduce((acc, cur) => acc + cur, 0) / datasets.length;
                }
                else {
                    min = datasets.reduce((acc, cur) => acc + cur, 0);
                }
            }
            if (Array.isArray(max)) {
                const datasets = max.map((r) => r.rp !== undefined ? r.rp : r.count);
                if (maxValueOperation === 'max') {
                    max = datasets.sort((a, b) => b - a)[0];
                }
                else if (maxValueOperation === 'min') {
                    max = datasets.sort((a, b) => a - b)[0];
                }
                else if (maxValueOperation === 'avg') {
                    max = datasets.reduce((acc, cur) => acc + cur, 0) / datasets.length;
                }
                else {
                    max = datasets.reduce((acc, cur) => acc + cur, 0);
                }
            }
            return { value, min, max };
        });
    }
}
Registry.set(kind$3, PageBlockProgress);

const defOptions$1 = {
    enabled: true,
    textColor: '#61AFFF',
    backgroundColor: '',
    item: {
        label: '',
        url: '',
        target: '',
        delimiter: false,
        pageID: '',
        pageLayoutID: '',
        moduleID: '',
        displaySubPages: false,
        align: 'bottom',
        dropdown: {
            label: '',
            items: [],
        },
    },
};
class NavigationItem {
    constructor(i) {
        this.type = '';
        this.options = Object.assign({}, defOptions$1);
        this.apply(i);
    }
    apply(i) {
        if (!i)
            return;
        Apply(this, i, String, 'type');
        if (i.options) {
            this.options = Object.assign(Object.assign({}, this.options), i.options);
        }
    }
}

const kind$2 = 'Navigation';
const defaults$2 = Object.freeze({
    display: {
        appearance: 'pills',
        alignment: 'center',
        justify: 'justify',
    },
    navigationItems: [],
    magnifyOption: '',
});
class PageBlockNavigation extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$2;
        this.options = Object.assign({}, defaults$2);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, String, 'magnifyOption');
        this.options.navigationItems = (o.navigationItems || []).map(f => new NavigationItem(f));
        this.options.display = Object.assign(Object.assign({}, this.options.display), o.display);
    }
    static makeNavigationItem(item) {
        return new NavigationItem(item);
    }
}
Registry.set(kind$2, PageBlockNavigation);

const kind$1 = 'Tabs';
const defaults$1 = Object.freeze({
    style: {
        appearance: 'tabs',
        alignment: 'center',
        justify: 'justify',
        orientation: 'horizontal',
        position: 'start',
    },
    tabs: [],
    magnifyOption: '',
});
class PageBlockTab extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind$1;
        this.options = Object.assign({}, defaults$1);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        Apply(this.options, o, String, 'magnifyOption');
        if (o.tabs) {
            this.options.tabs = o.tabs.map(tab => (Object.assign(Object.assign({}, tab), { lazy: tab.lazy !== false })));
        }
        if (o.style) {
            this.options.style = Object.assign(Object.assign({}, this.options.style), o.style);
        }
    }
}
Registry.set(kind$1, PageBlockTab);

const defOptions = {
    moduleID: NoID,
    color: '#61AFFF',
    prefilter: '',
};
/**
 * Feed class represents an event feed for the given calendar
 */
class Feed {
    constructor(i) {
        this.resource = 'compose:record';
        this.titleField = '';
        this.geometryField = '';
        this.displayMarker = true;
        this.displayPolygon = false;
        this.options = Object.assign({}, defOptions);
        this.apply(i);
    }
    apply(i) {
        if (!i)
            return;
        if (IsOf(i, 'resource')) {
            Apply(this, i, String, 'resource', 'titleField', 'geometryField');
            Apply(this, i, Boolean, 'displayMarker', 'displayPolygon');
            if (i.options) {
                this.options = Object.assign(Object.assign({}, this.options), i.options);
            }
        }
    }
    isValid() {
        return this.options.moduleID !== NoID && !!this.geometryField;
    }
}

function RecordFeed($ComposeAPI_1, module_1, namespace_1, feed_1) {
    return __awaiter(this, arguments, void 0, function* ($ComposeAPI, module, namespace, feed, options = {}) {
        // Params for record fetching
        const params = {
            namespaceID: namespace.namespaceID,
            moduleID: module.moduleID,
            query: feed.options.prefilter,
        };
        return $ComposeAPI.recordList(params, options).then(({ set }) => {
            return set
                // cast & freeze
                .map(r => Object.freeze(new Record(module, r)));
        });
    });
}

const kind = 'Geometry';
const defaults = Object.freeze({
    defaultView: '',
    center: [35, -30],
    feeds: [],
    zoomStarting: 2,
    zoomMin: 1,
    zoomMax: 18,
    bounds: null,
    lockBounds: false,
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
    displayOption: 'sameTab',
    hideGeoSearch: true,
});
class PageBlockGeometry extends PageBlock {
    constructor(i) {
        super(i);
        this.kind = kind;
        this.options = Object.assign({}, defaults);
        this.applyOptions(i === null || i === void 0 ? void 0 : i.options);
    }
    applyOptions(o) {
        if (!o)
            return;
        this.options.feeds = (o.feeds || []).map(f => new Feed(f));
        this.options.center = (o.center || []);
        this.options.bounds = (o.bounds || null);
        Apply(this.options, o, String, 'magnifyOption', 'displayOption');
        Apply(this.options, o, Number, 'zoomStarting', 'zoomMin', 'zoomMax', 'refreshRate');
        Apply(this.options, o, Boolean, 'lockBounds', 'showRefresh', 'hideGeoSearch');
    }
    static makeFeed(f) {
        return new Feed(f);
    }
}
PageBlockGeometry.feedResources = Object.freeze({
    record: 'compose:record',
});
PageBlockGeometry.RecordFeed = RecordFeed;
Registry.set(kind, PageBlockGeometry);

function PageBlockMaker(i) {
    const PageBlockTemp = Registry.get(i.kind);
    if (PageBlockTemp === undefined) {
        throw new Error(`unknown block kind '${i.kind}'`);
    }
    if (i instanceof PageBlock) {
        // Get rid of the references
        i = JSON.parse(JSON.stringify(i));
    }
    return new PageBlockTemp(i);
}

const { merge: merge$3 } = lodash;
class Page {
    constructor(i) {
        this.pageID = NoID;
        this.selfID = NoID;
        this.moduleID = NoID;
        this.namespaceID = NoID;
        this.title = '';
        this.handle = '';
        this.description = '';
        this.weight = 0;
        this.labels = {};
        this.visible = false;
        this.blocks = [];
        this.config = {
            navItem: {
                icon: {
                    type: '',
                    src: '',
                },
                expanded: false,
            },
        };
        this.meta = {
            notifications: {
                enabled: true,
            },
        };
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.canUpdatePage = false;
        this.canDeletePage = false;
        this.canGrant = false;
        this.apply(i);
    }
    clone() {
        return new Page(JSON.parse(JSON.stringify(this)));
    }
    apply(i) {
        if (!i)
            return;
        Apply(this, i, CortezaID, 'pageID', 'selfID', 'moduleID', 'namespaceID');
        Apply(this, i, String, 'title', 'handle', 'description');
        Apply(this, i, Number, 'weight');
        Apply(this, i, Boolean, 'visible');
        if (i.blocks) {
            this.blocks = i.blocks.map(block => PageBlockMaker(block));
        }
        if (i.children) {
            this.children = [];
            if (AreObjectsOf(i.children, 'pageID')) {
                this.children = i.children.map(c => new Page(c));
            }
        }
        if (IsOf(i, 'config')) {
            this.config = merge$3({}, this.config, i.config);
        }
        if (IsOf(i, 'meta')) {
            this.meta = merge$3({}, this.meta, i.meta);
        }
        if (IsOf(i, 'labels')) {
            this.labels = Object.assign({}, i.labels);
        }
        Apply(this, i, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, i, Boolean, 'canUpdatePage', 'canDeletePage', 'canGrant');
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.pageID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'compose:page';
    }
    get isRecordPage() {
        return this.moduleID !== NoID;
    }
    get firstLevel() {
        return this.selfID === NoID;
    }
    /**
     * Validates page & it's blocks
     */
    validate() {
        const ee = [];
        if (this.blocks.length === 0) {
            ee.push('blocks missing');
        }
        else {
            this.blocks.forEach(b => {
                ee.push(...b.validate());
            });
        }
        return ee;
    }
    export() {
        return {
            title: this.title,
            handle: this.handle,
            description: this.description,
            visible: this.visible,
            blocks: this.blocks,
        };
    }
}

const { merge: merge$2 } = lodash;
class PageLayout {
    constructor(pl) {
        this.pageLayoutID = NoID;
        this.namespaceID = NoID;
        this.pageID = NoID;
        this.handle = '';
        this.weight = 0;
        this.blocks = [];
        this.config = {
            visibility: {
                expression: '',
                roles: [],
            },
            buttons: {
                back: { enabled: true },
                delete: { enabled: true },
                clone: { enabled: true },
                new: { enabled: true },
                edit: { enabled: true },
                submit: { enabled: true },
            },
            actions: [],
            useTitle: false,
            validation: {
                requiredFields: [],
            },
        };
        this.meta = {
            title: '',
            description: '',
        };
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.ownedBy = NoID;
        this.apply(pl);
    }
    apply(pl) {
        if (!pl)
            return;
        Apply(this, pl, CortezaID, 'pageLayoutID', 'namespaceID', 'pageID', 'ownedBy');
        Apply(this, pl, String, 'handle');
        Apply(this, pl, Number, 'weight');
        Apply(this, pl, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        this.blocks = (pl.blocks || []).map(({ blockID, xywh, meta }) => ({ blockID, xywh, meta }));
        if (pl.meta) {
            this.meta = Object.assign(Object.assign({}, this.meta), pl.meta);
        }
        if (pl.config) {
            this.config = merge$2({}, this.config, pl.config);
        }
    }
    clone() {
        return new PageLayout(JSON.parse(JSON.stringify(this)));
    }
    addAction() {
        this.config.actions.push({
            kind: 'toLayout',
            placement: 'end',
            enabled: true,
            params: {
                pageLayoutID: '',
            },
            meta: {
                label: '',
                style: {
                    variant: 'primary',
                },
            },
        });
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.pageLayoutID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'compose:page-layout';
    }
    export() {
        return {
            blocks: this.blocks,
            meta: this.meta,
        };
    }
}

const { merge: merge$1 } = lodash;
class ValidatorError {
    constructor(i) {
        /**
         * Any additional meta data that can be used to expand (translated) message,
         * or to group, categorize validator results
         */
        this.meta = {};
        if (typeof i === 'string') {
            this.kind = i;
            this.message = i;
        }
        else {
            this.kind = i.kind;
            this.message = i.message || i.kind;
            if (i.meta) {
                this.meta = merge$1({}, this.meta, i.meta);
            }
        }
    }
}
const ValidatorFalseDefaultError = Object.freeze(new ValidatorError({
    message: 'Internal error',
    kind: 'internal',
}));
function NormalizeValidatorResults(...r) {
    const out = [];
    r.forEach(r => {
        if (r === undefined || r === null || r === true) {
            // considering these validation results as valid
            return;
        }
        if (Array.isArray(r)) {
            // Expand & normalize each item...
            out.push(...(NormalizeValidatorResults(...r)));
            return;
        }
        // eslint-disable-next-line @typescript-eslint/no-use-before-define
        if (r instanceof Validated) {
            out.push(...r.get());
            return;
        }
        if (r instanceof ValidatorError) {
            out.push(r);
            return;
        }
        if (IsOf(r, 'kind') && r.kind) {
            out.push(new ValidatorError(r));
            return;
        }
        // Catch-all for non object errors
        out.push(ValidatorFalseDefaultError);
    });
    return out;
}
/**
 * Holds an manipulates set of errors
 */
class Validated {
    constructor(...r) {
        this.set = [];
        this.push(...r);
    }
    get() {
        return this.set;
    }
    get length() {
        return this.set.length;
    }
    valid() {
        return this.length === 0;
    }
    push(...r) {
        this.set.push(...NormalizeValidatorResults(...r));
    }
    applyMeta(meta) {
        this.set = this.set.map(r => {
            const appliedMeta = Object.assign(Object.assign({}, r), { meta: Object.assign(Object.assign({}, r.meta), meta) });
            if (r instanceof ValidatorError) {
                return new ValidatorError(appliedMeta);
            }
            return new ValidatorError(appliedMeta);
        });
    }
    filter(fn) {
        return new Validated(this.set.filter(fn));
    }
    /**
     * Filters by meta keys
     *
     * If only key is given it returns entries that have meta with that key
     *
     * @param {string} key
     * @param {unknown} value
     */
    filterByMeta(key, value) {
        return this.filter((err) => (value === undefined ? err.meta[key] !== undefined : err.meta[key] === value));
    }
}
function IsEmpty(v) {
    if (!v || (IsOf(v, 'length') && v.length && v.length === 0)) {
        return true;
    }
    if (Array.isArray(v)) {
        return v.every(i => !i);
    }
    return !v;
}
/**
 * Checks if values are equal
 * @param {string|string[]} v1 Value in question
 * @param {string|string[]} v2 Value to compare to
 * @returns {boolean}
 */
function AreEqual(v1, v2) {
    if (Array.isArray(v1)) {
        if (!Array.isArray(v2) || v1.length !== v2.length) {
            return false;
        }
        return !!v1.find((v, i) => v !== v2[i]);
    }
    else {
        return v1 === v2;
    }
}
/**
 * Validator is record validation tool that registers and runs record & field validators
 *
 * Record and field validators are functions that
 */
class Validator {
    constructor(...vfn) {
        /**
         * Validators
         */
        this.registered = [];
        if (vfn) {
            this.registered.push(...vfn);
        }
    }
    push(...vfn) {
        this.registered.push(...vfn);
    }
    run(target, ...args) {
        return new Validated(...this.registered.map(vfn => vfn.call(target, ...args)));
    }
}

var validator = /*#__PURE__*/Object.freeze({
  __proto__: null,
  AreEqual: AreEqual,
  IsEmpty: IsEmpty,
  NormalizeValidatorResults: NormalizeValidatorResults,
  Validated: Validated,
  Validator: Validator,
  ValidatorError: ValidatorError
});

const emptyErr = new ValidatorError({ kind: 'empty', message: 'field:required-field' });
function genericFieldValidator(field) {
    // newValue is of type unknown to satisfy ValidatorFn interface
    return function (arg0) {
        if (!IsOf(arg0, 'field', 'value', 'oldValue')) {
            throw Error('invalid field validator argument type');
        }
        const { value } = arg0;
        if (field.isRequired) {
            const isNewRecord = this.recordID === NoID;
            const canManageFieldValue = isNewRecord ? true : field.canReadRecordValue && field.canUpdateRecordValue;
            if ((value === undefined || IsEmpty(value)) && canManageFieldValue) {
                return emptyErr;
            }
        }
    };
}
class RecordValidator extends Validator {
    /**
     * Construct record validator from module (or record)
     *
     * @param m
     */
    constructor(m) {
        super();
        this.rfv = {};
        if (m instanceof Record) {
            m = m.module;
        }
        m.fields.forEach(field => {
            this.rfv[field.name] = new Validator(genericFieldValidator(field));
        });
    }
    /**
     * Append more record validators
     *
     * @param name
     * @param vfn
     */
    push(...vfn) {
        this.registered.push(...vfn);
    }
    /**
     * Append more field validators
     *
     * @param name
     * @param vfn
     */
    pushToField(name, ...vfn) {
        if (!this.rfv[name]) {
            throw new Error('can not push validators to unknown field');
        }
        this.rfv[name].push(...vfn);
    }
    /**
     * Runs validators on record and all (or whitelisted) fields
     */
    run(r, ...fields) {
        const out = new Validated();
        if (fields.length === 0) {
            // Fields are not explicitly provided,
            // we can run record-wide validators:
            const result = super.run(r);
            out.push(result.get());
            // get list of fields from registered field validators
            fields = Object.getOwnPropertyNames(this.rfv);
        }
        for (const f of fields) {
            const field = r.module.findField(f);
            if (!field) {
                continue;
            }
            const payload = {
                value: r.values[f],
                oldValue: r.cleanValues[f],
                field,
            };
            const { recordID = NoID } = r || {};
            const results = this.rfv[f].run(r, payload);
            results.applyMeta({ field: f, id: recordID === NoID ? 'parent:0' : recordID });
            out.push(results.get());
        }
        return out;
    }
}

/*! js-yaml 4.1.0 https://github.com/nodeca/js-yaml @license MIT */
function isNothing(subject) {
  return (typeof subject === 'undefined') || (subject === null);
}


function isObject(subject) {
  return (typeof subject === 'object') && (subject !== null);
}


function toArray(sequence) {
  if (Array.isArray(sequence)) return sequence;
  else if (isNothing(sequence)) return [];

  return [ sequence ];
}


function extend(target, source) {
  var index, length, key, sourceKeys;

  if (source) {
    sourceKeys = Object.keys(source);

    for (index = 0, length = sourceKeys.length; index < length; index += 1) {
      key = sourceKeys[index];
      target[key] = source[key];
    }
  }

  return target;
}


function repeat(string, count) {
  var result = '', cycle;

  for (cycle = 0; cycle < count; cycle += 1) {
    result += string;
  }

  return result;
}


function isNegativeZero(number) {
  return (number === 0) && (Number.NEGATIVE_INFINITY === 1 / number);
}


var isNothing_1      = isNothing;
var isObject_1       = isObject;
var toArray_1        = toArray;
var repeat_1         = repeat;
var isNegativeZero_1 = isNegativeZero;
var extend_1         = extend;

var common = {
	isNothing: isNothing_1,
	isObject: isObject_1,
	toArray: toArray_1,
	repeat: repeat_1,
	isNegativeZero: isNegativeZero_1,
	extend: extend_1
};

// YAML error class. http://stackoverflow.com/questions/8458984


function formatError(exception, compact) {
  var where = '', message = exception.reason || '(unknown reason)';

  if (!exception.mark) return message;

  if (exception.mark.name) {
    where += 'in "' + exception.mark.name + '" ';
  }

  where += '(' + (exception.mark.line + 1) + ':' + (exception.mark.column + 1) + ')';

  if (!compact && exception.mark.snippet) {
    where += '\n\n' + exception.mark.snippet;
  }

  return message + ' ' + where;
}


function YAMLException$1(reason, mark) {
  // Super constructor
  Error.call(this);

  this.name = 'YAMLException';
  this.reason = reason;
  this.mark = mark;
  this.message = formatError(this, false);

  // Include stack trace in error object
  if (Error.captureStackTrace) {
    // Chrome and NodeJS
    Error.captureStackTrace(this, this.constructor);
  } else {
    // FF, IE 10+ and Safari 6+. Fallback for others
    this.stack = (new Error()).stack || '';
  }
}


// Inherit from Error
YAMLException$1.prototype = Object.create(Error.prototype);
YAMLException$1.prototype.constructor = YAMLException$1;


YAMLException$1.prototype.toString = function toString(compact) {
  return this.name + ': ' + formatError(this, compact);
};


var exception = YAMLException$1;

// get snippet for a single line, respecting maxLength
function getLine(buffer, lineStart, lineEnd, position, maxLineLength) {
  var head = '';
  var tail = '';
  var maxHalfLength = Math.floor(maxLineLength / 2) - 1;

  if (position - lineStart > maxHalfLength) {
    head = ' ... ';
    lineStart = position - maxHalfLength + head.length;
  }

  if (lineEnd - position > maxHalfLength) {
    tail = ' ...';
    lineEnd = position + maxHalfLength - tail.length;
  }

  return {
    str: head + buffer.slice(lineStart, lineEnd).replace(/\t/g, '→') + tail,
    pos: position - lineStart + head.length // relative position
  };
}


function padStart(string, max) {
  return common.repeat(' ', max - string.length) + string;
}


function makeSnippet(mark, options) {
  options = Object.create(options || null);

  if (!mark.buffer) return null;

  if (!options.maxLength) options.maxLength = 79;
  if (typeof options.indent      !== 'number') options.indent      = 1;
  if (typeof options.linesBefore !== 'number') options.linesBefore = 3;
  if (typeof options.linesAfter  !== 'number') options.linesAfter  = 2;

  var re = /\r?\n|\r|\0/g;
  var lineStarts = [ 0 ];
  var lineEnds = [];
  var match;
  var foundLineNo = -1;

  while ((match = re.exec(mark.buffer))) {
    lineEnds.push(match.index);
    lineStarts.push(match.index + match[0].length);

    if (mark.position <= match.index && foundLineNo < 0) {
      foundLineNo = lineStarts.length - 2;
    }
  }

  if (foundLineNo < 0) foundLineNo = lineStarts.length - 1;

  var result = '', i, line;
  var lineNoLength = Math.min(mark.line + options.linesAfter, lineEnds.length).toString().length;
  var maxLineLength = options.maxLength - (options.indent + lineNoLength + 3);

  for (i = 1; i <= options.linesBefore; i++) {
    if (foundLineNo - i < 0) break;
    line = getLine(
      mark.buffer,
      lineStarts[foundLineNo - i],
      lineEnds[foundLineNo - i],
      mark.position - (lineStarts[foundLineNo] - lineStarts[foundLineNo - i]),
      maxLineLength
    );
    result = common.repeat(' ', options.indent) + padStart((mark.line - i + 1).toString(), lineNoLength) +
      ' | ' + line.str + '\n' + result;
  }

  line = getLine(mark.buffer, lineStarts[foundLineNo], lineEnds[foundLineNo], mark.position, maxLineLength);
  result += common.repeat(' ', options.indent) + padStart((mark.line + 1).toString(), lineNoLength) +
    ' | ' + line.str + '\n';
  result += common.repeat('-', options.indent + lineNoLength + 3 + line.pos) + '^' + '\n';

  for (i = 1; i <= options.linesAfter; i++) {
    if (foundLineNo + i >= lineEnds.length) break;
    line = getLine(
      mark.buffer,
      lineStarts[foundLineNo + i],
      lineEnds[foundLineNo + i],
      mark.position - (lineStarts[foundLineNo] - lineStarts[foundLineNo + i]),
      maxLineLength
    );
    result += common.repeat(' ', options.indent) + padStart((mark.line + i + 1).toString(), lineNoLength) +
      ' | ' + line.str + '\n';
  }

  return result.replace(/\n$/, '');
}


var snippet = makeSnippet;

var TYPE_CONSTRUCTOR_OPTIONS = [
  'kind',
  'multi',
  'resolve',
  'construct',
  'instanceOf',
  'predicate',
  'represent',
  'representName',
  'defaultStyle',
  'styleAliases'
];

var YAML_NODE_KINDS = [
  'scalar',
  'sequence',
  'mapping'
];

function compileStyleAliases(map) {
  var result = {};

  if (map !== null) {
    Object.keys(map).forEach(function (style) {
      map[style].forEach(function (alias) {
        result[String(alias)] = style;
      });
    });
  }

  return result;
}

function Type$1(tag, options) {
  options = options || {};

  Object.keys(options).forEach(function (name) {
    if (TYPE_CONSTRUCTOR_OPTIONS.indexOf(name) === -1) {
      throw new exception('Unknown option "' + name + '" is met in definition of "' + tag + '" YAML type.');
    }
  });

  // TODO: Add tag format check.
  this.options       = options; // keep original options in case user wants to extend this type later
  this.tag           = tag;
  this.kind          = options['kind']          || null;
  this.resolve       = options['resolve']       || function () { return true; };
  this.construct     = options['construct']     || function (data) { return data; };
  this.instanceOf    = options['instanceOf']    || null;
  this.predicate     = options['predicate']     || null;
  this.represent     = options['represent']     || null;
  this.representName = options['representName'] || null;
  this.defaultStyle  = options['defaultStyle']  || null;
  this.multi         = options['multi']         || false;
  this.styleAliases  = compileStyleAliases(options['styleAliases'] || null);

  if (YAML_NODE_KINDS.indexOf(this.kind) === -1) {
    throw new exception('Unknown kind "' + this.kind + '" is specified for "' + tag + '" YAML type.');
  }
}

var type = Type$1;

/*eslint-disable max-len*/





function compileList(schema, name) {
  var result = [];

  schema[name].forEach(function (currentType) {
    var newIndex = result.length;

    result.forEach(function (previousType, previousIndex) {
      if (previousType.tag === currentType.tag &&
          previousType.kind === currentType.kind &&
          previousType.multi === currentType.multi) {

        newIndex = previousIndex;
      }
    });

    result[newIndex] = currentType;
  });

  return result;
}


function compileMap(/* lists... */) {
  var result = {
        scalar: {},
        sequence: {},
        mapping: {},
        fallback: {},
        multi: {
          scalar: [],
          sequence: [],
          mapping: [],
          fallback: []
        }
      }, index, length;

  function collectType(type) {
    if (type.multi) {
      result.multi[type.kind].push(type);
      result.multi['fallback'].push(type);
    } else {
      result[type.kind][type.tag] = result['fallback'][type.tag] = type;
    }
  }

  for (index = 0, length = arguments.length; index < length; index += 1) {
    arguments[index].forEach(collectType);
  }
  return result;
}


function Schema$1(definition) {
  return this.extend(definition);
}


Schema$1.prototype.extend = function extend(definition) {
  var implicit = [];
  var explicit = [];

  if (definition instanceof type) {
    // Schema.extend(type)
    explicit.push(definition);

  } else if (Array.isArray(definition)) {
    // Schema.extend([ type1, type2, ... ])
    explicit = explicit.concat(definition);

  } else if (definition && (Array.isArray(definition.implicit) || Array.isArray(definition.explicit))) {
    // Schema.extend({ explicit: [ type1, type2, ... ], implicit: [ type1, type2, ... ] })
    if (definition.implicit) implicit = implicit.concat(definition.implicit);
    if (definition.explicit) explicit = explicit.concat(definition.explicit);

  } else {
    throw new exception('Schema.extend argument should be a Type, [ Type ], ' +
      'or a schema definition ({ implicit: [...], explicit: [...] })');
  }

  implicit.forEach(function (type$1) {
    if (!(type$1 instanceof type)) {
      throw new exception('Specified list of YAML types (or a single Type object) contains a non-Type object.');
    }

    if (type$1.loadKind && type$1.loadKind !== 'scalar') {
      throw new exception('There is a non-scalar type in the implicit list of a schema. Implicit resolving of such types is not supported.');
    }

    if (type$1.multi) {
      throw new exception('There is a multi type in the implicit list of a schema. Multi tags can only be listed as explicit.');
    }
  });

  explicit.forEach(function (type$1) {
    if (!(type$1 instanceof type)) {
      throw new exception('Specified list of YAML types (or a single Type object) contains a non-Type object.');
    }
  });

  var result = Object.create(Schema$1.prototype);

  result.implicit = (this.implicit || []).concat(implicit);
  result.explicit = (this.explicit || []).concat(explicit);

  result.compiledImplicit = compileList(result, 'implicit');
  result.compiledExplicit = compileList(result, 'explicit');
  result.compiledTypeMap  = compileMap(result.compiledImplicit, result.compiledExplicit);

  return result;
};


var schema = Schema$1;

var str = new type('tag:yaml.org,2002:str', {
  kind: 'scalar',
  construct: function (data) { return data !== null ? data : ''; }
});

var seq = new type('tag:yaml.org,2002:seq', {
  kind: 'sequence',
  construct: function (data) { return data !== null ? data : []; }
});

var map = new type('tag:yaml.org,2002:map', {
  kind: 'mapping',
  construct: function (data) { return data !== null ? data : {}; }
});

var failsafe = new schema({
  explicit: [
    str,
    seq,
    map
  ]
});

function resolveYamlNull(data) {
  if (data === null) return true;

  var max = data.length;

  return (max === 1 && data === '~') ||
         (max === 4 && (data === 'null' || data === 'Null' || data === 'NULL'));
}

function constructYamlNull() {
  return null;
}

function isNull(object) {
  return object === null;
}

var _null = new type('tag:yaml.org,2002:null', {
  kind: 'scalar',
  resolve: resolveYamlNull,
  construct: constructYamlNull,
  predicate: isNull,
  represent: {
    canonical: function () { return '~';    },
    lowercase: function () { return 'null'; },
    uppercase: function () { return 'NULL'; },
    camelcase: function () { return 'Null'; },
    empty:     function () { return '';     }
  },
  defaultStyle: 'lowercase'
});

function resolveYamlBoolean(data) {
  if (data === null) return false;

  var max = data.length;

  return (max === 4 && (data === 'true' || data === 'True' || data === 'TRUE')) ||
         (max === 5 && (data === 'false' || data === 'False' || data === 'FALSE'));
}

function constructYamlBoolean(data) {
  return data === 'true' ||
         data === 'True' ||
         data === 'TRUE';
}

function isBoolean(object) {
  return Object.prototype.toString.call(object) === '[object Boolean]';
}

var bool = new type('tag:yaml.org,2002:bool', {
  kind: 'scalar',
  resolve: resolveYamlBoolean,
  construct: constructYamlBoolean,
  predicate: isBoolean,
  represent: {
    lowercase: function (object) { return object ? 'true' : 'false'; },
    uppercase: function (object) { return object ? 'TRUE' : 'FALSE'; },
    camelcase: function (object) { return object ? 'True' : 'False'; }
  },
  defaultStyle: 'lowercase'
});

function isHexCode(c) {
  return ((0x30/* 0 */ <= c) && (c <= 0x39/* 9 */)) ||
         ((0x41/* A */ <= c) && (c <= 0x46/* F */)) ||
         ((0x61/* a */ <= c) && (c <= 0x66/* f */));
}

function isOctCode(c) {
  return ((0x30/* 0 */ <= c) && (c <= 0x37/* 7 */));
}

function isDecCode(c) {
  return ((0x30/* 0 */ <= c) && (c <= 0x39/* 9 */));
}

function resolveYamlInteger(data) {
  if (data === null) return false;

  var max = data.length,
      index = 0,
      hasDigits = false,
      ch;

  if (!max) return false;

  ch = data[index];

  // sign
  if (ch === '-' || ch === '+') {
    ch = data[++index];
  }

  if (ch === '0') {
    // 0
    if (index + 1 === max) return true;
    ch = data[++index];

    // base 2, base 8, base 16

    if (ch === 'b') {
      // base 2
      index++;

      for (; index < max; index++) {
        ch = data[index];
        if (ch === '_') continue;
        if (ch !== '0' && ch !== '1') return false;
        hasDigits = true;
      }
      return hasDigits && ch !== '_';
    }


    if (ch === 'x') {
      // base 16
      index++;

      for (; index < max; index++) {
        ch = data[index];
        if (ch === '_') continue;
        if (!isHexCode(data.charCodeAt(index))) return false;
        hasDigits = true;
      }
      return hasDigits && ch !== '_';
    }


    if (ch === 'o') {
      // base 8
      index++;

      for (; index < max; index++) {
        ch = data[index];
        if (ch === '_') continue;
        if (!isOctCode(data.charCodeAt(index))) return false;
        hasDigits = true;
      }
      return hasDigits && ch !== '_';
    }
  }

  // base 10 (except 0)

  // value should not start with `_`;
  if (ch === '_') return false;

  for (; index < max; index++) {
    ch = data[index];
    if (ch === '_') continue;
    if (!isDecCode(data.charCodeAt(index))) {
      return false;
    }
    hasDigits = true;
  }

  // Should have digits and should not end with `_`
  if (!hasDigits || ch === '_') return false;

  return true;
}

function constructYamlInteger(data) {
  var value = data, sign = 1, ch;

  if (value.indexOf('_') !== -1) {
    value = value.replace(/_/g, '');
  }

  ch = value[0];

  if (ch === '-' || ch === '+') {
    if (ch === '-') sign = -1;
    value = value.slice(1);
    ch = value[0];
  }

  if (value === '0') return 0;

  if (ch === '0') {
    if (value[1] === 'b') return sign * parseInt(value.slice(2), 2);
    if (value[1] === 'x') return sign * parseInt(value.slice(2), 16);
    if (value[1] === 'o') return sign * parseInt(value.slice(2), 8);
  }

  return sign * parseInt(value, 10);
}

function isInteger(object) {
  return (Object.prototype.toString.call(object)) === '[object Number]' &&
         (object % 1 === 0 && !common.isNegativeZero(object));
}

var int = new type('tag:yaml.org,2002:int', {
  kind: 'scalar',
  resolve: resolveYamlInteger,
  construct: constructYamlInteger,
  predicate: isInteger,
  represent: {
    binary:      function (obj) { return obj >= 0 ? '0b' + obj.toString(2) : '-0b' + obj.toString(2).slice(1); },
    octal:       function (obj) { return obj >= 0 ? '0o'  + obj.toString(8) : '-0o'  + obj.toString(8).slice(1); },
    decimal:     function (obj) { return obj.toString(10); },
    /* eslint-disable max-len */
    hexadecimal: function (obj) { return obj >= 0 ? '0x' + obj.toString(16).toUpperCase() :  '-0x' + obj.toString(16).toUpperCase().slice(1); }
  },
  defaultStyle: 'decimal',
  styleAliases: {
    binary:      [ 2,  'bin' ],
    octal:       [ 8,  'oct' ],
    decimal:     [ 10, 'dec' ],
    hexadecimal: [ 16, 'hex' ]
  }
});

var YAML_FLOAT_PATTERN = new RegExp(
  // 2.5e4, 2.5 and integers
  '^(?:[-+]?(?:[0-9][0-9_]*)(?:\\.[0-9_]*)?(?:[eE][-+]?[0-9]+)?' +
  // .2e4, .2
  // special case, seems not from spec
  '|\\.[0-9_]+(?:[eE][-+]?[0-9]+)?' +
  // .inf
  '|[-+]?\\.(?:inf|Inf|INF)' +
  // .nan
  '|\\.(?:nan|NaN|NAN))$');

function resolveYamlFloat(data) {
  if (data === null) return false;

  if (!YAML_FLOAT_PATTERN.test(data) ||
      // Quick hack to not allow integers end with `_`
      // Probably should update regexp & check speed
      data[data.length - 1] === '_') {
    return false;
  }

  return true;
}

function constructYamlFloat(data) {
  var value, sign;

  value  = data.replace(/_/g, '').toLowerCase();
  sign   = value[0] === '-' ? -1 : 1;

  if ('+-'.indexOf(value[0]) >= 0) {
    value = value.slice(1);
  }

  if (value === '.inf') {
    return (sign === 1) ? Number.POSITIVE_INFINITY : Number.NEGATIVE_INFINITY;

  } else if (value === '.nan') {
    return NaN;
  }
  return sign * parseFloat(value, 10);
}


var SCIENTIFIC_WITHOUT_DOT = /^[-+]?[0-9]+e/;

function representYamlFloat(object, style) {
  var res;

  if (isNaN(object)) {
    switch (style) {
      case 'lowercase': return '.nan';
      case 'uppercase': return '.NAN';
      case 'camelcase': return '.NaN';
    }
  } else if (Number.POSITIVE_INFINITY === object) {
    switch (style) {
      case 'lowercase': return '.inf';
      case 'uppercase': return '.INF';
      case 'camelcase': return '.Inf';
    }
  } else if (Number.NEGATIVE_INFINITY === object) {
    switch (style) {
      case 'lowercase': return '-.inf';
      case 'uppercase': return '-.INF';
      case 'camelcase': return '-.Inf';
    }
  } else if (common.isNegativeZero(object)) {
    return '-0.0';
  }

  res = object.toString(10);

  // JS stringifier can build scientific format without dots: 5e-100,
  // while YAML requres dot: 5.e-100. Fix it with simple hack

  return SCIENTIFIC_WITHOUT_DOT.test(res) ? res.replace('e', '.e') : res;
}

function isFloat(object) {
  return (Object.prototype.toString.call(object) === '[object Number]') &&
         (object % 1 !== 0 || common.isNegativeZero(object));
}

var float = new type('tag:yaml.org,2002:float', {
  kind: 'scalar',
  resolve: resolveYamlFloat,
  construct: constructYamlFloat,
  predicate: isFloat,
  represent: representYamlFloat,
  defaultStyle: 'lowercase'
});

var json = failsafe.extend({
  implicit: [
    _null,
    bool,
    int,
    float
  ]
});

var core = json;

var YAML_DATE_REGEXP = new RegExp(
  '^([0-9][0-9][0-9][0-9])'          + // [1] year
  '-([0-9][0-9])'                    + // [2] month
  '-([0-9][0-9])$');                   // [3] day

var YAML_TIMESTAMP_REGEXP = new RegExp(
  '^([0-9][0-9][0-9][0-9])'          + // [1] year
  '-([0-9][0-9]?)'                   + // [2] month
  '-([0-9][0-9]?)'                   + // [3] day
  '(?:[Tt]|[ \\t]+)'                 + // ...
  '([0-9][0-9]?)'                    + // [4] hour
  ':([0-9][0-9])'                    + // [5] minute
  ':([0-9][0-9])'                    + // [6] second
  '(?:\\.([0-9]*))?'                 + // [7] fraction
  '(?:[ \\t]*(Z|([-+])([0-9][0-9]?)' + // [8] tz [9] tz_sign [10] tz_hour
  '(?::([0-9][0-9]))?))?$');           // [11] tz_minute

function resolveYamlTimestamp(data) {
  if (data === null) return false;
  if (YAML_DATE_REGEXP.exec(data) !== null) return true;
  if (YAML_TIMESTAMP_REGEXP.exec(data) !== null) return true;
  return false;
}

function constructYamlTimestamp(data) {
  var match, year, month, day, hour, minute, second, fraction = 0,
      delta = null, tz_hour, tz_minute, date;

  match = YAML_DATE_REGEXP.exec(data);
  if (match === null) match = YAML_TIMESTAMP_REGEXP.exec(data);

  if (match === null) throw new Error('Date resolve error');

  // match: [1] year [2] month [3] day

  year = +(match[1]);
  month = +(match[2]) - 1; // JS month starts with 0
  day = +(match[3]);

  if (!match[4]) { // no hour
    return new Date(Date.UTC(year, month, day));
  }

  // match: [4] hour [5] minute [6] second [7] fraction

  hour = +(match[4]);
  minute = +(match[5]);
  second = +(match[6]);

  if (match[7]) {
    fraction = match[7].slice(0, 3);
    while (fraction.length < 3) { // milli-seconds
      fraction += '0';
    }
    fraction = +fraction;
  }

  // match: [8] tz [9] tz_sign [10] tz_hour [11] tz_minute

  if (match[9]) {
    tz_hour = +(match[10]);
    tz_minute = +(match[11] || 0);
    delta = (tz_hour * 60 + tz_minute) * 60000; // delta in mili-seconds
    if (match[9] === '-') delta = -delta;
  }

  date = new Date(Date.UTC(year, month, day, hour, minute, second, fraction));

  if (delta) date.setTime(date.getTime() - delta);

  return date;
}

function representYamlTimestamp(object /*, style*/) {
  return object.toISOString();
}

var timestamp = new type('tag:yaml.org,2002:timestamp', {
  kind: 'scalar',
  resolve: resolveYamlTimestamp,
  construct: constructYamlTimestamp,
  instanceOf: Date,
  represent: representYamlTimestamp
});

function resolveYamlMerge(data) {
  return data === '<<' || data === null;
}

var merge = new type('tag:yaml.org,2002:merge', {
  kind: 'scalar',
  resolve: resolveYamlMerge
});

/*eslint-disable no-bitwise*/





// [ 64, 65, 66 ] -> [ padding, CR, LF ]
var BASE64_MAP = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=\n\r';


function resolveYamlBinary(data) {
  if (data === null) return false;

  var code, idx, bitlen = 0, max = data.length, map = BASE64_MAP;

  // Convert one by one.
  for (idx = 0; idx < max; idx++) {
    code = map.indexOf(data.charAt(idx));

    // Skip CR/LF
    if (code > 64) continue;

    // Fail on illegal characters
    if (code < 0) return false;

    bitlen += 6;
  }

  // If there are any bits left, source was corrupted
  return (bitlen % 8) === 0;
}

function constructYamlBinary(data) {
  var idx, tailbits,
      input = data.replace(/[\r\n=]/g, ''), // remove CR/LF & padding to simplify scan
      max = input.length,
      map = BASE64_MAP,
      bits = 0,
      result = [];

  // Collect by 6*4 bits (3 bytes)

  for (idx = 0; idx < max; idx++) {
    if ((idx % 4 === 0) && idx) {
      result.push((bits >> 16) & 0xFF);
      result.push((bits >> 8) & 0xFF);
      result.push(bits & 0xFF);
    }

    bits = (bits << 6) | map.indexOf(input.charAt(idx));
  }

  // Dump tail

  tailbits = (max % 4) * 6;

  if (tailbits === 0) {
    result.push((bits >> 16) & 0xFF);
    result.push((bits >> 8) & 0xFF);
    result.push(bits & 0xFF);
  } else if (tailbits === 18) {
    result.push((bits >> 10) & 0xFF);
    result.push((bits >> 2) & 0xFF);
  } else if (tailbits === 12) {
    result.push((bits >> 4) & 0xFF);
  }

  return new Uint8Array(result);
}

function representYamlBinary(object /*, style*/) {
  var result = '', bits = 0, idx, tail,
      max = object.length,
      map = BASE64_MAP;

  // Convert every three bytes to 4 ASCII characters.

  for (idx = 0; idx < max; idx++) {
    if ((idx % 3 === 0) && idx) {
      result += map[(bits >> 18) & 0x3F];
      result += map[(bits >> 12) & 0x3F];
      result += map[(bits >> 6) & 0x3F];
      result += map[bits & 0x3F];
    }

    bits = (bits << 8) + object[idx];
  }

  // Dump tail

  tail = max % 3;

  if (tail === 0) {
    result += map[(bits >> 18) & 0x3F];
    result += map[(bits >> 12) & 0x3F];
    result += map[(bits >> 6) & 0x3F];
    result += map[bits & 0x3F];
  } else if (tail === 2) {
    result += map[(bits >> 10) & 0x3F];
    result += map[(bits >> 4) & 0x3F];
    result += map[(bits << 2) & 0x3F];
    result += map[64];
  } else if (tail === 1) {
    result += map[(bits >> 2) & 0x3F];
    result += map[(bits << 4) & 0x3F];
    result += map[64];
    result += map[64];
  }

  return result;
}

function isBinary(obj) {
  return Object.prototype.toString.call(obj) ===  '[object Uint8Array]';
}

var binary = new type('tag:yaml.org,2002:binary', {
  kind: 'scalar',
  resolve: resolveYamlBinary,
  construct: constructYamlBinary,
  predicate: isBinary,
  represent: representYamlBinary
});

var _hasOwnProperty$3 = Object.prototype.hasOwnProperty;
var _toString$2       = Object.prototype.toString;

function resolveYamlOmap(data) {
  if (data === null) return true;

  var objectKeys = [], index, length, pair, pairKey, pairHasKey,
      object = data;

  for (index = 0, length = object.length; index < length; index += 1) {
    pair = object[index];
    pairHasKey = false;

    if (_toString$2.call(pair) !== '[object Object]') return false;

    for (pairKey in pair) {
      if (_hasOwnProperty$3.call(pair, pairKey)) {
        if (!pairHasKey) pairHasKey = true;
        else return false;
      }
    }

    if (!pairHasKey) return false;

    if (objectKeys.indexOf(pairKey) === -1) objectKeys.push(pairKey);
    else return false;
  }

  return true;
}

function constructYamlOmap(data) {
  return data !== null ? data : [];
}

var omap = new type('tag:yaml.org,2002:omap', {
  kind: 'sequence',
  resolve: resolveYamlOmap,
  construct: constructYamlOmap
});

var _toString$1 = Object.prototype.toString;

function resolveYamlPairs(data) {
  if (data === null) return true;

  var index, length, pair, keys, result,
      object = data;

  result = new Array(object.length);

  for (index = 0, length = object.length; index < length; index += 1) {
    pair = object[index];

    if (_toString$1.call(pair) !== '[object Object]') return false;

    keys = Object.keys(pair);

    if (keys.length !== 1) return false;

    result[index] = [ keys[0], pair[keys[0]] ];
  }

  return true;
}

function constructYamlPairs(data) {
  if (data === null) return [];

  var index, length, pair, keys, result,
      object = data;

  result = new Array(object.length);

  for (index = 0, length = object.length; index < length; index += 1) {
    pair = object[index];

    keys = Object.keys(pair);

    result[index] = [ keys[0], pair[keys[0]] ];
  }

  return result;
}

var pairs = new type('tag:yaml.org,2002:pairs', {
  kind: 'sequence',
  resolve: resolveYamlPairs,
  construct: constructYamlPairs
});

var _hasOwnProperty$2 = Object.prototype.hasOwnProperty;

function resolveYamlSet(data) {
  if (data === null) return true;

  var key, object = data;

  for (key in object) {
    if (_hasOwnProperty$2.call(object, key)) {
      if (object[key] !== null) return false;
    }
  }

  return true;
}

function constructYamlSet(data) {
  return data !== null ? data : {};
}

var set = new type('tag:yaml.org,2002:set', {
  kind: 'mapping',
  resolve: resolveYamlSet,
  construct: constructYamlSet
});

var _default = core.extend({
  implicit: [
    timestamp,
    merge
  ],
  explicit: [
    binary,
    omap,
    pairs,
    set
  ]
});

/*eslint-disable max-len,no-use-before-define*/







var _hasOwnProperty$1 = Object.prototype.hasOwnProperty;


var CONTEXT_FLOW_IN   = 1;
var CONTEXT_FLOW_OUT  = 2;
var CONTEXT_BLOCK_IN  = 3;
var CONTEXT_BLOCK_OUT = 4;


var CHOMPING_CLIP  = 1;
var CHOMPING_STRIP = 2;
var CHOMPING_KEEP  = 3;


var PATTERN_NON_PRINTABLE         = /[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x84\x86-\x9F\uFFFE\uFFFF]|[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?:[^\uD800-\uDBFF]|^)[\uDC00-\uDFFF]/;
var PATTERN_NON_ASCII_LINE_BREAKS = /[\x85\u2028\u2029]/;
var PATTERN_FLOW_INDICATORS       = /[,\[\]\{\}]/;
var PATTERN_TAG_HANDLE            = /^(?:!|!!|![a-z\-]+!)$/i;
var PATTERN_TAG_URI               = /^(?:!|[^,\[\]\{\}])(?:%[0-9a-f]{2}|[0-9a-z\-#;\/\?:@&=\+\$,_\.!~\*'\(\)\[\]])*$/i;


function _class(obj) { return Object.prototype.toString.call(obj); }

function is_EOL(c) {
  return (c === 0x0A/* LF */) || (c === 0x0D/* CR */);
}

function is_WHITE_SPACE(c) {
  return (c === 0x09/* Tab */) || (c === 0x20/* Space */);
}

function is_WS_OR_EOL(c) {
  return (c === 0x09/* Tab */) ||
         (c === 0x20/* Space */) ||
         (c === 0x0A/* LF */) ||
         (c === 0x0D/* CR */);
}

function is_FLOW_INDICATOR(c) {
  return c === 0x2C/* , */ ||
         c === 0x5B/* [ */ ||
         c === 0x5D/* ] */ ||
         c === 0x7B/* { */ ||
         c === 0x7D/* } */;
}

function fromHexCode(c) {
  var lc;

  if ((0x30/* 0 */ <= c) && (c <= 0x39/* 9 */)) {
    return c - 0x30;
  }

  /*eslint-disable no-bitwise*/
  lc = c | 0x20;

  if ((0x61/* a */ <= lc) && (lc <= 0x66/* f */)) {
    return lc - 0x61 + 10;
  }

  return -1;
}

function escapedHexLen(c) {
  if (c === 0x78/* x */) { return 2; }
  if (c === 0x75/* u */) { return 4; }
  if (c === 0x55/* U */) { return 8; }
  return 0;
}

function fromDecimalCode(c) {
  if ((0x30/* 0 */ <= c) && (c <= 0x39/* 9 */)) {
    return c - 0x30;
  }

  return -1;
}

function simpleEscapeSequence(c) {
  /* eslint-disable indent */
  return (c === 0x30/* 0 */) ? '\x00' :
        (c === 0x61/* a */) ? '\x07' :
        (c === 0x62/* b */) ? '\x08' :
        (c === 0x74/* t */) ? '\x09' :
        (c === 0x09/* Tab */) ? '\x09' :
        (c === 0x6E/* n */) ? '\x0A' :
        (c === 0x76/* v */) ? '\x0B' :
        (c === 0x66/* f */) ? '\x0C' :
        (c === 0x72/* r */) ? '\x0D' :
        (c === 0x65/* e */) ? '\x1B' :
        (c === 0x20/* Space */) ? ' ' :
        (c === 0x22/* " */) ? '\x22' :
        (c === 0x2F/* / */) ? '/' :
        (c === 0x5C/* \ */) ? '\x5C' :
        (c === 0x4E/* N */) ? '\x85' :
        (c === 0x5F/* _ */) ? '\xA0' :
        (c === 0x4C/* L */) ? '\u2028' :
        (c === 0x50/* P */) ? '\u2029' : '';
}

function charFromCodepoint(c) {
  if (c <= 0xFFFF) {
    return String.fromCharCode(c);
  }
  // Encode UTF-16 surrogate pair
  // https://en.wikipedia.org/wiki/UTF-16#Code_points_U.2B010000_to_U.2B10FFFF
  return String.fromCharCode(
    ((c - 0x010000) >> 10) + 0xD800,
    ((c - 0x010000) & 0x03FF) + 0xDC00
  );
}

var simpleEscapeCheck = new Array(256); // integer, for fast access
var simpleEscapeMap = new Array(256);
for (var i = 0; i < 256; i++) {
  simpleEscapeCheck[i] = simpleEscapeSequence(i) ? 1 : 0;
  simpleEscapeMap[i] = simpleEscapeSequence(i);
}


function State$1(input, options) {
  this.input = input;

  this.filename  = options['filename']  || null;
  this.schema    = options['schema']    || _default;
  this.onWarning = options['onWarning'] || null;
  // (Hidden) Remove? makes the loader to expect YAML 1.1 documents
  // if such documents have no explicit %YAML directive
  this.legacy    = options['legacy']    || false;

  this.json      = options['json']      || false;
  this.listener  = options['listener']  || null;

  this.implicitTypes = this.schema.compiledImplicit;
  this.typeMap       = this.schema.compiledTypeMap;

  this.length     = input.length;
  this.position   = 0;
  this.line       = 0;
  this.lineStart  = 0;
  this.lineIndent = 0;

  // position of first leading tab in the current line,
  // used to make sure there are no tabs in the indentation
  this.firstTabInLine = -1;

  this.documents = [];

  /*
  this.version;
  this.checkLineBreaks;
  this.tagMap;
  this.anchorMap;
  this.tag;
  this.anchor;
  this.kind;
  this.result;*/

}


function generateError(state, message) {
  var mark = {
    name:     state.filename,
    buffer:   state.input.slice(0, -1), // omit trailing \0
    position: state.position,
    line:     state.line,
    column:   state.position - state.lineStart
  };

  mark.snippet = snippet(mark);

  return new exception(message, mark);
}

function throwError(state, message) {
  throw generateError(state, message);
}

function throwWarning(state, message) {
  if (state.onWarning) {
    state.onWarning.call(null, generateError(state, message));
  }
}


var directiveHandlers = {

  YAML: function handleYamlDirective(state, name, args) {

    var match, major, minor;

    if (state.version !== null) {
      throwError(state, 'duplication of %YAML directive');
    }

    if (args.length !== 1) {
      throwError(state, 'YAML directive accepts exactly one argument');
    }

    match = /^([0-9]+)\.([0-9]+)$/.exec(args[0]);

    if (match === null) {
      throwError(state, 'ill-formed argument of the YAML directive');
    }

    major = parseInt(match[1], 10);
    minor = parseInt(match[2], 10);

    if (major !== 1) {
      throwError(state, 'unacceptable YAML version of the document');
    }

    state.version = args[0];
    state.checkLineBreaks = (minor < 2);

    if (minor !== 1 && minor !== 2) {
      throwWarning(state, 'unsupported YAML version of the document');
    }
  },

  TAG: function handleTagDirective(state, name, args) {

    var handle, prefix;

    if (args.length !== 2) {
      throwError(state, 'TAG directive accepts exactly two arguments');
    }

    handle = args[0];
    prefix = args[1];

    if (!PATTERN_TAG_HANDLE.test(handle)) {
      throwError(state, 'ill-formed tag handle (first argument) of the TAG directive');
    }

    if (_hasOwnProperty$1.call(state.tagMap, handle)) {
      throwError(state, 'there is a previously declared suffix for "' + handle + '" tag handle');
    }

    if (!PATTERN_TAG_URI.test(prefix)) {
      throwError(state, 'ill-formed tag prefix (second argument) of the TAG directive');
    }

    try {
      prefix = decodeURIComponent(prefix);
    } catch (err) {
      throwError(state, 'tag prefix is malformed: ' + prefix);
    }

    state.tagMap[handle] = prefix;
  }
};


function captureSegment(state, start, end, checkJson) {
  var _position, _length, _character, _result;

  if (start < end) {
    _result = state.input.slice(start, end);

    if (checkJson) {
      for (_position = 0, _length = _result.length; _position < _length; _position += 1) {
        _character = _result.charCodeAt(_position);
        if (!(_character === 0x09 ||
              (0x20 <= _character && _character <= 0x10FFFF))) {
          throwError(state, 'expected valid JSON character');
        }
      }
    } else if (PATTERN_NON_PRINTABLE.test(_result)) {
      throwError(state, 'the stream contains non-printable characters');
    }

    state.result += _result;
  }
}

function mergeMappings(state, destination, source, overridableKeys) {
  var sourceKeys, key, index, quantity;

  if (!common.isObject(source)) {
    throwError(state, 'cannot merge mappings; the provided source object is unacceptable');
  }

  sourceKeys = Object.keys(source);

  for (index = 0, quantity = sourceKeys.length; index < quantity; index += 1) {
    key = sourceKeys[index];

    if (!_hasOwnProperty$1.call(destination, key)) {
      destination[key] = source[key];
      overridableKeys[key] = true;
    }
  }
}

function storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode,
  startLine, startLineStart, startPos) {

  var index, quantity;

  // The output is a plain object here, so keys can only be strings.
  // We need to convert keyNode to a string, but doing so can hang the process
  // (deeply nested arrays that explode exponentially using aliases).
  if (Array.isArray(keyNode)) {
    keyNode = Array.prototype.slice.call(keyNode);

    for (index = 0, quantity = keyNode.length; index < quantity; index += 1) {
      if (Array.isArray(keyNode[index])) {
        throwError(state, 'nested arrays are not supported inside keys');
      }

      if (typeof keyNode === 'object' && _class(keyNode[index]) === '[object Object]') {
        keyNode[index] = '[object Object]';
      }
    }
  }

  // Avoid code execution in load() via toString property
  // (still use its own toString for arrays, timestamps,
  // and whatever user schema extensions happen to have @@toStringTag)
  if (typeof keyNode === 'object' && _class(keyNode) === '[object Object]') {
    keyNode = '[object Object]';
  }


  keyNode = String(keyNode);

  if (_result === null) {
    _result = {};
  }

  if (keyTag === 'tag:yaml.org,2002:merge') {
    if (Array.isArray(valueNode)) {
      for (index = 0, quantity = valueNode.length; index < quantity; index += 1) {
        mergeMappings(state, _result, valueNode[index], overridableKeys);
      }
    } else {
      mergeMappings(state, _result, valueNode, overridableKeys);
    }
  } else {
    if (!state.json &&
        !_hasOwnProperty$1.call(overridableKeys, keyNode) &&
        _hasOwnProperty$1.call(_result, keyNode)) {
      state.line = startLine || state.line;
      state.lineStart = startLineStart || state.lineStart;
      state.position = startPos || state.position;
      throwError(state, 'duplicated mapping key');
    }

    // used for this specific key only because Object.defineProperty is slow
    if (keyNode === '__proto__') {
      Object.defineProperty(_result, keyNode, {
        configurable: true,
        enumerable: true,
        writable: true,
        value: valueNode
      });
    } else {
      _result[keyNode] = valueNode;
    }
    delete overridableKeys[keyNode];
  }

  return _result;
}

function readLineBreak(state) {
  var ch;

  ch = state.input.charCodeAt(state.position);

  if (ch === 0x0A/* LF */) {
    state.position++;
  } else if (ch === 0x0D/* CR */) {
    state.position++;
    if (state.input.charCodeAt(state.position) === 0x0A/* LF */) {
      state.position++;
    }
  } else {
    throwError(state, 'a line break is expected');
  }

  state.line += 1;
  state.lineStart = state.position;
  state.firstTabInLine = -1;
}

function skipSeparationSpace(state, allowComments, checkIndent) {
  var lineBreaks = 0,
      ch = state.input.charCodeAt(state.position);

  while (ch !== 0) {
    while (is_WHITE_SPACE(ch)) {
      if (ch === 0x09/* Tab */ && state.firstTabInLine === -1) {
        state.firstTabInLine = state.position;
      }
      ch = state.input.charCodeAt(++state.position);
    }

    if (allowComments && ch === 0x23/* # */) {
      do {
        ch = state.input.charCodeAt(++state.position);
      } while (ch !== 0x0A/* LF */ && ch !== 0x0D/* CR */ && ch !== 0);
    }

    if (is_EOL(ch)) {
      readLineBreak(state);

      ch = state.input.charCodeAt(state.position);
      lineBreaks++;
      state.lineIndent = 0;

      while (ch === 0x20/* Space */) {
        state.lineIndent++;
        ch = state.input.charCodeAt(++state.position);
      }
    } else {
      break;
    }
  }

  if (checkIndent !== -1 && lineBreaks !== 0 && state.lineIndent < checkIndent) {
    throwWarning(state, 'deficient indentation');
  }

  return lineBreaks;
}

function testDocumentSeparator(state) {
  var _position = state.position,
      ch;

  ch = state.input.charCodeAt(_position);

  // Condition state.position === state.lineStart is tested
  // in parent on each call, for efficiency. No needs to test here again.
  if ((ch === 0x2D/* - */ || ch === 0x2E/* . */) &&
      ch === state.input.charCodeAt(_position + 1) &&
      ch === state.input.charCodeAt(_position + 2)) {

    _position += 3;

    ch = state.input.charCodeAt(_position);

    if (ch === 0 || is_WS_OR_EOL(ch)) {
      return true;
    }
  }

  return false;
}

function writeFoldedLines(state, count) {
  if (count === 1) {
    state.result += ' ';
  } else if (count > 1) {
    state.result += common.repeat('\n', count - 1);
  }
}


function readPlainScalar(state, nodeIndent, withinFlowCollection) {
  var preceding,
      following,
      captureStart,
      captureEnd,
      hasPendingContent,
      _line,
      _lineStart,
      _lineIndent,
      _kind = state.kind,
      _result = state.result,
      ch;

  ch = state.input.charCodeAt(state.position);

  if (is_WS_OR_EOL(ch)      ||
      is_FLOW_INDICATOR(ch) ||
      ch === 0x23/* # */    ||
      ch === 0x26/* & */    ||
      ch === 0x2A/* * */    ||
      ch === 0x21/* ! */    ||
      ch === 0x7C/* | */    ||
      ch === 0x3E/* > */    ||
      ch === 0x27/* ' */    ||
      ch === 0x22/* " */    ||
      ch === 0x25/* % */    ||
      ch === 0x40/* @ */    ||
      ch === 0x60/* ` */) {
    return false;
  }

  if (ch === 0x3F/* ? */ || ch === 0x2D/* - */) {
    following = state.input.charCodeAt(state.position + 1);

    if (is_WS_OR_EOL(following) ||
        withinFlowCollection && is_FLOW_INDICATOR(following)) {
      return false;
    }
  }

  state.kind = 'scalar';
  state.result = '';
  captureStart = captureEnd = state.position;
  hasPendingContent = false;

  while (ch !== 0) {
    if (ch === 0x3A/* : */) {
      following = state.input.charCodeAt(state.position + 1);

      if (is_WS_OR_EOL(following) ||
          withinFlowCollection && is_FLOW_INDICATOR(following)) {
        break;
      }

    } else if (ch === 0x23/* # */) {
      preceding = state.input.charCodeAt(state.position - 1);

      if (is_WS_OR_EOL(preceding)) {
        break;
      }

    } else if ((state.position === state.lineStart && testDocumentSeparator(state)) ||
               withinFlowCollection && is_FLOW_INDICATOR(ch)) {
      break;

    } else if (is_EOL(ch)) {
      _line = state.line;
      _lineStart = state.lineStart;
      _lineIndent = state.lineIndent;
      skipSeparationSpace(state, false, -1);

      if (state.lineIndent >= nodeIndent) {
        hasPendingContent = true;
        ch = state.input.charCodeAt(state.position);
        continue;
      } else {
        state.position = captureEnd;
        state.line = _line;
        state.lineStart = _lineStart;
        state.lineIndent = _lineIndent;
        break;
      }
    }

    if (hasPendingContent) {
      captureSegment(state, captureStart, captureEnd, false);
      writeFoldedLines(state, state.line - _line);
      captureStart = captureEnd = state.position;
      hasPendingContent = false;
    }

    if (!is_WHITE_SPACE(ch)) {
      captureEnd = state.position + 1;
    }

    ch = state.input.charCodeAt(++state.position);
  }

  captureSegment(state, captureStart, captureEnd, false);

  if (state.result) {
    return true;
  }

  state.kind = _kind;
  state.result = _result;
  return false;
}

function readSingleQuotedScalar(state, nodeIndent) {
  var ch,
      captureStart, captureEnd;

  ch = state.input.charCodeAt(state.position);

  if (ch !== 0x27/* ' */) {
    return false;
  }

  state.kind = 'scalar';
  state.result = '';
  state.position++;
  captureStart = captureEnd = state.position;

  while ((ch = state.input.charCodeAt(state.position)) !== 0) {
    if (ch === 0x27/* ' */) {
      captureSegment(state, captureStart, state.position, true);
      ch = state.input.charCodeAt(++state.position);

      if (ch === 0x27/* ' */) {
        captureStart = state.position;
        state.position++;
        captureEnd = state.position;
      } else {
        return true;
      }

    } else if (is_EOL(ch)) {
      captureSegment(state, captureStart, captureEnd, true);
      writeFoldedLines(state, skipSeparationSpace(state, false, nodeIndent));
      captureStart = captureEnd = state.position;

    } else if (state.position === state.lineStart && testDocumentSeparator(state)) {
      throwError(state, 'unexpected end of the document within a single quoted scalar');

    } else {
      state.position++;
      captureEnd = state.position;
    }
  }

  throwError(state, 'unexpected end of the stream within a single quoted scalar');
}

function readDoubleQuotedScalar(state, nodeIndent) {
  var captureStart,
      captureEnd,
      hexLength,
      hexResult,
      tmp,
      ch;

  ch = state.input.charCodeAt(state.position);

  if (ch !== 0x22/* " */) {
    return false;
  }

  state.kind = 'scalar';
  state.result = '';
  state.position++;
  captureStart = captureEnd = state.position;

  while ((ch = state.input.charCodeAt(state.position)) !== 0) {
    if (ch === 0x22/* " */) {
      captureSegment(state, captureStart, state.position, true);
      state.position++;
      return true;

    } else if (ch === 0x5C/* \ */) {
      captureSegment(state, captureStart, state.position, true);
      ch = state.input.charCodeAt(++state.position);

      if (is_EOL(ch)) {
        skipSeparationSpace(state, false, nodeIndent);

        // TODO: rework to inline fn with no type cast?
      } else if (ch < 256 && simpleEscapeCheck[ch]) {
        state.result += simpleEscapeMap[ch];
        state.position++;

      } else if ((tmp = escapedHexLen(ch)) > 0) {
        hexLength = tmp;
        hexResult = 0;

        for (; hexLength > 0; hexLength--) {
          ch = state.input.charCodeAt(++state.position);

          if ((tmp = fromHexCode(ch)) >= 0) {
            hexResult = (hexResult << 4) + tmp;

          } else {
            throwError(state, 'expected hexadecimal character');
          }
        }

        state.result += charFromCodepoint(hexResult);

        state.position++;

      } else {
        throwError(state, 'unknown escape sequence');
      }

      captureStart = captureEnd = state.position;

    } else if (is_EOL(ch)) {
      captureSegment(state, captureStart, captureEnd, true);
      writeFoldedLines(state, skipSeparationSpace(state, false, nodeIndent));
      captureStart = captureEnd = state.position;

    } else if (state.position === state.lineStart && testDocumentSeparator(state)) {
      throwError(state, 'unexpected end of the document within a double quoted scalar');

    } else {
      state.position++;
      captureEnd = state.position;
    }
  }

  throwError(state, 'unexpected end of the stream within a double quoted scalar');
}

function readFlowCollection(state, nodeIndent) {
  var readNext = true,
      _line,
      _lineStart,
      _pos,
      _tag     = state.tag,
      _result,
      _anchor  = state.anchor,
      following,
      terminator,
      isPair,
      isExplicitPair,
      isMapping,
      overridableKeys = Object.create(null),
      keyNode,
      keyTag,
      valueNode,
      ch;

  ch = state.input.charCodeAt(state.position);

  if (ch === 0x5B/* [ */) {
    terminator = 0x5D;/* ] */
    isMapping = false;
    _result = [];
  } else if (ch === 0x7B/* { */) {
    terminator = 0x7D;/* } */
    isMapping = true;
    _result = {};
  } else {
    return false;
  }

  if (state.anchor !== null) {
    state.anchorMap[state.anchor] = _result;
  }

  ch = state.input.charCodeAt(++state.position);

  while (ch !== 0) {
    skipSeparationSpace(state, true, nodeIndent);

    ch = state.input.charCodeAt(state.position);

    if (ch === terminator) {
      state.position++;
      state.tag = _tag;
      state.anchor = _anchor;
      state.kind = isMapping ? 'mapping' : 'sequence';
      state.result = _result;
      return true;
    } else if (!readNext) {
      throwError(state, 'missed comma between flow collection entries');
    } else if (ch === 0x2C/* , */) {
      // "flow collection entries can never be completely empty", as per YAML 1.2, section 7.4
      throwError(state, "expected the node content, but found ','");
    }

    keyTag = keyNode = valueNode = null;
    isPair = isExplicitPair = false;

    if (ch === 0x3F/* ? */) {
      following = state.input.charCodeAt(state.position + 1);

      if (is_WS_OR_EOL(following)) {
        isPair = isExplicitPair = true;
        state.position++;
        skipSeparationSpace(state, true, nodeIndent);
      }
    }

    _line = state.line; // Save the current line.
    _lineStart = state.lineStart;
    _pos = state.position;
    composeNode(state, nodeIndent, CONTEXT_FLOW_IN, false, true);
    keyTag = state.tag;
    keyNode = state.result;
    skipSeparationSpace(state, true, nodeIndent);

    ch = state.input.charCodeAt(state.position);

    if ((isExplicitPair || state.line === _line) && ch === 0x3A/* : */) {
      isPair = true;
      ch = state.input.charCodeAt(++state.position);
      skipSeparationSpace(state, true, nodeIndent);
      composeNode(state, nodeIndent, CONTEXT_FLOW_IN, false, true);
      valueNode = state.result;
    }

    if (isMapping) {
      storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, _line, _lineStart, _pos);
    } else if (isPair) {
      _result.push(storeMappingPair(state, null, overridableKeys, keyTag, keyNode, valueNode, _line, _lineStart, _pos));
    } else {
      _result.push(keyNode);
    }

    skipSeparationSpace(state, true, nodeIndent);

    ch = state.input.charCodeAt(state.position);

    if (ch === 0x2C/* , */) {
      readNext = true;
      ch = state.input.charCodeAt(++state.position);
    } else {
      readNext = false;
    }
  }

  throwError(state, 'unexpected end of the stream within a flow collection');
}

function readBlockScalar(state, nodeIndent) {
  var captureStart,
      folding,
      chomping       = CHOMPING_CLIP,
      didReadContent = false,
      detectedIndent = false,
      textIndent     = nodeIndent,
      emptyLines     = 0,
      atMoreIndented = false,
      tmp,
      ch;

  ch = state.input.charCodeAt(state.position);

  if (ch === 0x7C/* | */) {
    folding = false;
  } else if (ch === 0x3E/* > */) {
    folding = true;
  } else {
    return false;
  }

  state.kind = 'scalar';
  state.result = '';

  while (ch !== 0) {
    ch = state.input.charCodeAt(++state.position);

    if (ch === 0x2B/* + */ || ch === 0x2D/* - */) {
      if (CHOMPING_CLIP === chomping) {
        chomping = (ch === 0x2B/* + */) ? CHOMPING_KEEP : CHOMPING_STRIP;
      } else {
        throwError(state, 'repeat of a chomping mode identifier');
      }

    } else if ((tmp = fromDecimalCode(ch)) >= 0) {
      if (tmp === 0) {
        throwError(state, 'bad explicit indentation width of a block scalar; it cannot be less than one');
      } else if (!detectedIndent) {
        textIndent = nodeIndent + tmp - 1;
        detectedIndent = true;
      } else {
        throwError(state, 'repeat of an indentation width identifier');
      }

    } else {
      break;
    }
  }

  if (is_WHITE_SPACE(ch)) {
    do { ch = state.input.charCodeAt(++state.position); }
    while (is_WHITE_SPACE(ch));

    if (ch === 0x23/* # */) {
      do { ch = state.input.charCodeAt(++state.position); }
      while (!is_EOL(ch) && (ch !== 0));
    }
  }

  while (ch !== 0) {
    readLineBreak(state);
    state.lineIndent = 0;

    ch = state.input.charCodeAt(state.position);

    while ((!detectedIndent || state.lineIndent < textIndent) &&
           (ch === 0x20/* Space */)) {
      state.lineIndent++;
      ch = state.input.charCodeAt(++state.position);
    }

    if (!detectedIndent && state.lineIndent > textIndent) {
      textIndent = state.lineIndent;
    }

    if (is_EOL(ch)) {
      emptyLines++;
      continue;
    }

    // End of the scalar.
    if (state.lineIndent < textIndent) {

      // Perform the chomping.
      if (chomping === CHOMPING_KEEP) {
        state.result += common.repeat('\n', didReadContent ? 1 + emptyLines : emptyLines);
      } else if (chomping === CHOMPING_CLIP) {
        if (didReadContent) { // i.e. only if the scalar is not empty.
          state.result += '\n';
        }
      }

      // Break this `while` cycle and go to the funciton's epilogue.
      break;
    }

    // Folded style: use fancy rules to handle line breaks.
    if (folding) {

      // Lines starting with white space characters (more-indented lines) are not folded.
      if (is_WHITE_SPACE(ch)) {
        atMoreIndented = true;
        // except for the first content line (cf. Example 8.1)
        state.result += common.repeat('\n', didReadContent ? 1 + emptyLines : emptyLines);

      // End of more-indented block.
      } else if (atMoreIndented) {
        atMoreIndented = false;
        state.result += common.repeat('\n', emptyLines + 1);

      // Just one line break - perceive as the same line.
      } else if (emptyLines === 0) {
        if (didReadContent) { // i.e. only if we have already read some scalar content.
          state.result += ' ';
        }

      // Several line breaks - perceive as different lines.
      } else {
        state.result += common.repeat('\n', emptyLines);
      }

    // Literal style: just add exact number of line breaks between content lines.
    } else {
      // Keep all line breaks except the header line break.
      state.result += common.repeat('\n', didReadContent ? 1 + emptyLines : emptyLines);
    }

    didReadContent = true;
    detectedIndent = true;
    emptyLines = 0;
    captureStart = state.position;

    while (!is_EOL(ch) && (ch !== 0)) {
      ch = state.input.charCodeAt(++state.position);
    }

    captureSegment(state, captureStart, state.position, false);
  }

  return true;
}

function readBlockSequence(state, nodeIndent) {
  var _line,
      _tag      = state.tag,
      _anchor   = state.anchor,
      _result   = [],
      following,
      detected  = false,
      ch;

  // there is a leading tab before this token, so it can't be a block sequence/mapping;
  // it can still be flow sequence/mapping or a scalar
  if (state.firstTabInLine !== -1) return false;

  if (state.anchor !== null) {
    state.anchorMap[state.anchor] = _result;
  }

  ch = state.input.charCodeAt(state.position);

  while (ch !== 0) {
    if (state.firstTabInLine !== -1) {
      state.position = state.firstTabInLine;
      throwError(state, 'tab characters must not be used in indentation');
    }

    if (ch !== 0x2D/* - */) {
      break;
    }

    following = state.input.charCodeAt(state.position + 1);

    if (!is_WS_OR_EOL(following)) {
      break;
    }

    detected = true;
    state.position++;

    if (skipSeparationSpace(state, true, -1)) {
      if (state.lineIndent <= nodeIndent) {
        _result.push(null);
        ch = state.input.charCodeAt(state.position);
        continue;
      }
    }

    _line = state.line;
    composeNode(state, nodeIndent, CONTEXT_BLOCK_IN, false, true);
    _result.push(state.result);
    skipSeparationSpace(state, true, -1);

    ch = state.input.charCodeAt(state.position);

    if ((state.line === _line || state.lineIndent > nodeIndent) && (ch !== 0)) {
      throwError(state, 'bad indentation of a sequence entry');
    } else if (state.lineIndent < nodeIndent) {
      break;
    }
  }

  if (detected) {
    state.tag = _tag;
    state.anchor = _anchor;
    state.kind = 'sequence';
    state.result = _result;
    return true;
  }
  return false;
}

function readBlockMapping(state, nodeIndent, flowIndent) {
  var following,
      allowCompact,
      _line,
      _keyLine,
      _keyLineStart,
      _keyPos,
      _tag          = state.tag,
      _anchor       = state.anchor,
      _result       = {},
      overridableKeys = Object.create(null),
      keyTag        = null,
      keyNode       = null,
      valueNode     = null,
      atExplicitKey = false,
      detected      = false,
      ch;

  // there is a leading tab before this token, so it can't be a block sequence/mapping;
  // it can still be flow sequence/mapping or a scalar
  if (state.firstTabInLine !== -1) return false;

  if (state.anchor !== null) {
    state.anchorMap[state.anchor] = _result;
  }

  ch = state.input.charCodeAt(state.position);

  while (ch !== 0) {
    if (!atExplicitKey && state.firstTabInLine !== -1) {
      state.position = state.firstTabInLine;
      throwError(state, 'tab characters must not be used in indentation');
    }

    following = state.input.charCodeAt(state.position + 1);
    _line = state.line; // Save the current line.

    //
    // Explicit notation case. There are two separate blocks:
    // first for the key (denoted by "?") and second for the value (denoted by ":")
    //
    if ((ch === 0x3F/* ? */ || ch === 0x3A/* : */) && is_WS_OR_EOL(following)) {

      if (ch === 0x3F/* ? */) {
        if (atExplicitKey) {
          storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
          keyTag = keyNode = valueNode = null;
        }

        detected = true;
        atExplicitKey = true;
        allowCompact = true;

      } else if (atExplicitKey) {
        // i.e. 0x3A/* : */ === character after the explicit key.
        atExplicitKey = false;
        allowCompact = true;

      } else {
        throwError(state, 'incomplete explicit mapping pair; a key node is missed; or followed by a non-tabulated empty line');
      }

      state.position += 1;
      ch = following;

    //
    // Implicit notation case. Flow-style node as the key first, then ":", and the value.
    //
    } else {
      _keyLine = state.line;
      _keyLineStart = state.lineStart;
      _keyPos = state.position;

      if (!composeNode(state, flowIndent, CONTEXT_FLOW_OUT, false, true)) {
        // Neither implicit nor explicit notation.
        // Reading is done. Go to the epilogue.
        break;
      }

      if (state.line === _line) {
        ch = state.input.charCodeAt(state.position);

        while (is_WHITE_SPACE(ch)) {
          ch = state.input.charCodeAt(++state.position);
        }

        if (ch === 0x3A/* : */) {
          ch = state.input.charCodeAt(++state.position);

          if (!is_WS_OR_EOL(ch)) {
            throwError(state, 'a whitespace character is expected after the key-value separator within a block mapping');
          }

          if (atExplicitKey) {
            storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
            keyTag = keyNode = valueNode = null;
          }

          detected = true;
          atExplicitKey = false;
          allowCompact = false;
          keyTag = state.tag;
          keyNode = state.result;

        } else if (detected) {
          throwError(state, 'can not read an implicit mapping pair; a colon is missed');

        } else {
          state.tag = _tag;
          state.anchor = _anchor;
          return true; // Keep the result of `composeNode`.
        }

      } else if (detected) {
        throwError(state, 'can not read a block mapping entry; a multiline key may not be an implicit key');

      } else {
        state.tag = _tag;
        state.anchor = _anchor;
        return true; // Keep the result of `composeNode`.
      }
    }

    //
    // Common reading code for both explicit and implicit notations.
    //
    if (state.line === _line || state.lineIndent > nodeIndent) {
      if (atExplicitKey) {
        _keyLine = state.line;
        _keyLineStart = state.lineStart;
        _keyPos = state.position;
      }

      if (composeNode(state, nodeIndent, CONTEXT_BLOCK_OUT, true, allowCompact)) {
        if (atExplicitKey) {
          keyNode = state.result;
        } else {
          valueNode = state.result;
        }
      }

      if (!atExplicitKey) {
        storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, _keyLine, _keyLineStart, _keyPos);
        keyTag = keyNode = valueNode = null;
      }

      skipSeparationSpace(state, true, -1);
      ch = state.input.charCodeAt(state.position);
    }

    if ((state.line === _line || state.lineIndent > nodeIndent) && (ch !== 0)) {
      throwError(state, 'bad indentation of a mapping entry');
    } else if (state.lineIndent < nodeIndent) {
      break;
    }
  }

  //
  // Epilogue.
  //

  // Special case: last mapping's node contains only the key in explicit notation.
  if (atExplicitKey) {
    storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
  }

  // Expose the resulting mapping.
  if (detected) {
    state.tag = _tag;
    state.anchor = _anchor;
    state.kind = 'mapping';
    state.result = _result;
  }

  return detected;
}

function readTagProperty(state) {
  var _position,
      isVerbatim = false,
      isNamed    = false,
      tagHandle,
      tagName,
      ch;

  ch = state.input.charCodeAt(state.position);

  if (ch !== 0x21/* ! */) return false;

  if (state.tag !== null) {
    throwError(state, 'duplication of a tag property');
  }

  ch = state.input.charCodeAt(++state.position);

  if (ch === 0x3C/* < */) {
    isVerbatim = true;
    ch = state.input.charCodeAt(++state.position);

  } else if (ch === 0x21/* ! */) {
    isNamed = true;
    tagHandle = '!!';
    ch = state.input.charCodeAt(++state.position);

  } else {
    tagHandle = '!';
  }

  _position = state.position;

  if (isVerbatim) {
    do { ch = state.input.charCodeAt(++state.position); }
    while (ch !== 0 && ch !== 0x3E/* > */);

    if (state.position < state.length) {
      tagName = state.input.slice(_position, state.position);
      ch = state.input.charCodeAt(++state.position);
    } else {
      throwError(state, 'unexpected end of the stream within a verbatim tag');
    }
  } else {
    while (ch !== 0 && !is_WS_OR_EOL(ch)) {

      if (ch === 0x21/* ! */) {
        if (!isNamed) {
          tagHandle = state.input.slice(_position - 1, state.position + 1);

          if (!PATTERN_TAG_HANDLE.test(tagHandle)) {
            throwError(state, 'named tag handle cannot contain such characters');
          }

          isNamed = true;
          _position = state.position + 1;
        } else {
          throwError(state, 'tag suffix cannot contain exclamation marks');
        }
      }

      ch = state.input.charCodeAt(++state.position);
    }

    tagName = state.input.slice(_position, state.position);

    if (PATTERN_FLOW_INDICATORS.test(tagName)) {
      throwError(state, 'tag suffix cannot contain flow indicator characters');
    }
  }

  if (tagName && !PATTERN_TAG_URI.test(tagName)) {
    throwError(state, 'tag name cannot contain such characters: ' + tagName);
  }

  try {
    tagName = decodeURIComponent(tagName);
  } catch (err) {
    throwError(state, 'tag name is malformed: ' + tagName);
  }

  if (isVerbatim) {
    state.tag = tagName;

  } else if (_hasOwnProperty$1.call(state.tagMap, tagHandle)) {
    state.tag = state.tagMap[tagHandle] + tagName;

  } else if (tagHandle === '!') {
    state.tag = '!' + tagName;

  } else if (tagHandle === '!!') {
    state.tag = 'tag:yaml.org,2002:' + tagName;

  } else {
    throwError(state, 'undeclared tag handle "' + tagHandle + '"');
  }

  return true;
}

function readAnchorProperty(state) {
  var _position,
      ch;

  ch = state.input.charCodeAt(state.position);

  if (ch !== 0x26/* & */) return false;

  if (state.anchor !== null) {
    throwError(state, 'duplication of an anchor property');
  }

  ch = state.input.charCodeAt(++state.position);
  _position = state.position;

  while (ch !== 0 && !is_WS_OR_EOL(ch) && !is_FLOW_INDICATOR(ch)) {
    ch = state.input.charCodeAt(++state.position);
  }

  if (state.position === _position) {
    throwError(state, 'name of an anchor node must contain at least one character');
  }

  state.anchor = state.input.slice(_position, state.position);
  return true;
}

function readAlias(state) {
  var _position, alias,
      ch;

  ch = state.input.charCodeAt(state.position);

  if (ch !== 0x2A/* * */) return false;

  ch = state.input.charCodeAt(++state.position);
  _position = state.position;

  while (ch !== 0 && !is_WS_OR_EOL(ch) && !is_FLOW_INDICATOR(ch)) {
    ch = state.input.charCodeAt(++state.position);
  }

  if (state.position === _position) {
    throwError(state, 'name of an alias node must contain at least one character');
  }

  alias = state.input.slice(_position, state.position);

  if (!_hasOwnProperty$1.call(state.anchorMap, alias)) {
    throwError(state, 'unidentified alias "' + alias + '"');
  }

  state.result = state.anchorMap[alias];
  skipSeparationSpace(state, true, -1);
  return true;
}

function composeNode(state, parentIndent, nodeContext, allowToSeek, allowCompact) {
  var allowBlockStyles,
      allowBlockScalars,
      allowBlockCollections,
      indentStatus = 1, // 1: this>parent, 0: this=parent, -1: this<parent
      atNewLine  = false,
      hasContent = false,
      typeIndex,
      typeQuantity,
      typeList,
      type,
      flowIndent,
      blockIndent;

  if (state.listener !== null) {
    state.listener('open', state);
  }

  state.tag    = null;
  state.anchor = null;
  state.kind   = null;
  state.result = null;

  allowBlockStyles = allowBlockScalars = allowBlockCollections =
    CONTEXT_BLOCK_OUT === nodeContext ||
    CONTEXT_BLOCK_IN  === nodeContext;

  if (allowToSeek) {
    if (skipSeparationSpace(state, true, -1)) {
      atNewLine = true;

      if (state.lineIndent > parentIndent) {
        indentStatus = 1;
      } else if (state.lineIndent === parentIndent) {
        indentStatus = 0;
      } else if (state.lineIndent < parentIndent) {
        indentStatus = -1;
      }
    }
  }

  if (indentStatus === 1) {
    while (readTagProperty(state) || readAnchorProperty(state)) {
      if (skipSeparationSpace(state, true, -1)) {
        atNewLine = true;
        allowBlockCollections = allowBlockStyles;

        if (state.lineIndent > parentIndent) {
          indentStatus = 1;
        } else if (state.lineIndent === parentIndent) {
          indentStatus = 0;
        } else if (state.lineIndent < parentIndent) {
          indentStatus = -1;
        }
      } else {
        allowBlockCollections = false;
      }
    }
  }

  if (allowBlockCollections) {
    allowBlockCollections = atNewLine || allowCompact;
  }

  if (indentStatus === 1 || CONTEXT_BLOCK_OUT === nodeContext) {
    if (CONTEXT_FLOW_IN === nodeContext || CONTEXT_FLOW_OUT === nodeContext) {
      flowIndent = parentIndent;
    } else {
      flowIndent = parentIndent + 1;
    }

    blockIndent = state.position - state.lineStart;

    if (indentStatus === 1) {
      if (allowBlockCollections &&
          (readBlockSequence(state, blockIndent) ||
           readBlockMapping(state, blockIndent, flowIndent)) ||
          readFlowCollection(state, flowIndent)) {
        hasContent = true;
      } else {
        if ((allowBlockScalars && readBlockScalar(state, flowIndent)) ||
            readSingleQuotedScalar(state, flowIndent) ||
            readDoubleQuotedScalar(state, flowIndent)) {
          hasContent = true;

        } else if (readAlias(state)) {
          hasContent = true;

          if (state.tag !== null || state.anchor !== null) {
            throwError(state, 'alias node should not have any properties');
          }

        } else if (readPlainScalar(state, flowIndent, CONTEXT_FLOW_IN === nodeContext)) {
          hasContent = true;

          if (state.tag === null) {
            state.tag = '?';
          }
        }

        if (state.anchor !== null) {
          state.anchorMap[state.anchor] = state.result;
        }
      }
    } else if (indentStatus === 0) {
      // Special case: block sequences are allowed to have same indentation level as the parent.
      // http://www.yaml.org/spec/1.2/spec.html#id2799784
      hasContent = allowBlockCollections && readBlockSequence(state, blockIndent);
    }
  }

  if (state.tag === null) {
    if (state.anchor !== null) {
      state.anchorMap[state.anchor] = state.result;
    }

  } else if (state.tag === '?') {
    // Implicit resolving is not allowed for non-scalar types, and '?'
    // non-specific tag is only automatically assigned to plain scalars.
    //
    // We only need to check kind conformity in case user explicitly assigns '?'
    // tag, for example like this: "!<?> [0]"
    //
    if (state.result !== null && state.kind !== 'scalar') {
      throwError(state, 'unacceptable node kind for !<?> tag; it should be "scalar", not "' + state.kind + '"');
    }

    for (typeIndex = 0, typeQuantity = state.implicitTypes.length; typeIndex < typeQuantity; typeIndex += 1) {
      type = state.implicitTypes[typeIndex];

      if (type.resolve(state.result)) { // `state.result` updated in resolver if matched
        state.result = type.construct(state.result);
        state.tag = type.tag;
        if (state.anchor !== null) {
          state.anchorMap[state.anchor] = state.result;
        }
        break;
      }
    }
  } else if (state.tag !== '!') {
    if (_hasOwnProperty$1.call(state.typeMap[state.kind || 'fallback'], state.tag)) {
      type = state.typeMap[state.kind || 'fallback'][state.tag];
    } else {
      // looking for multi type
      type = null;
      typeList = state.typeMap.multi[state.kind || 'fallback'];

      for (typeIndex = 0, typeQuantity = typeList.length; typeIndex < typeQuantity; typeIndex += 1) {
        if (state.tag.slice(0, typeList[typeIndex].tag.length) === typeList[typeIndex].tag) {
          type = typeList[typeIndex];
          break;
        }
      }
    }

    if (!type) {
      throwError(state, 'unknown tag !<' + state.tag + '>');
    }

    if (state.result !== null && type.kind !== state.kind) {
      throwError(state, 'unacceptable node kind for !<' + state.tag + '> tag; it should be "' + type.kind + '", not "' + state.kind + '"');
    }

    if (!type.resolve(state.result, state.tag)) { // `state.result` updated in resolver if matched
      throwError(state, 'cannot resolve a node with !<' + state.tag + '> explicit tag');
    } else {
      state.result = type.construct(state.result, state.tag);
      if (state.anchor !== null) {
        state.anchorMap[state.anchor] = state.result;
      }
    }
  }

  if (state.listener !== null) {
    state.listener('close', state);
  }
  return state.tag !== null ||  state.anchor !== null || hasContent;
}

function readDocument(state) {
  var documentStart = state.position,
      _position,
      directiveName,
      directiveArgs,
      hasDirectives = false,
      ch;

  state.version = null;
  state.checkLineBreaks = state.legacy;
  state.tagMap = Object.create(null);
  state.anchorMap = Object.create(null);

  while ((ch = state.input.charCodeAt(state.position)) !== 0) {
    skipSeparationSpace(state, true, -1);

    ch = state.input.charCodeAt(state.position);

    if (state.lineIndent > 0 || ch !== 0x25/* % */) {
      break;
    }

    hasDirectives = true;
    ch = state.input.charCodeAt(++state.position);
    _position = state.position;

    while (ch !== 0 && !is_WS_OR_EOL(ch)) {
      ch = state.input.charCodeAt(++state.position);
    }

    directiveName = state.input.slice(_position, state.position);
    directiveArgs = [];

    if (directiveName.length < 1) {
      throwError(state, 'directive name must not be less than one character in length');
    }

    while (ch !== 0) {
      while (is_WHITE_SPACE(ch)) {
        ch = state.input.charCodeAt(++state.position);
      }

      if (ch === 0x23/* # */) {
        do { ch = state.input.charCodeAt(++state.position); }
        while (ch !== 0 && !is_EOL(ch));
        break;
      }

      if (is_EOL(ch)) break;

      _position = state.position;

      while (ch !== 0 && !is_WS_OR_EOL(ch)) {
        ch = state.input.charCodeAt(++state.position);
      }

      directiveArgs.push(state.input.slice(_position, state.position));
    }

    if (ch !== 0) readLineBreak(state);

    if (_hasOwnProperty$1.call(directiveHandlers, directiveName)) {
      directiveHandlers[directiveName](state, directiveName, directiveArgs);
    } else {
      throwWarning(state, 'unknown document directive "' + directiveName + '"');
    }
  }

  skipSeparationSpace(state, true, -1);

  if (state.lineIndent === 0 &&
      state.input.charCodeAt(state.position)     === 0x2D/* - */ &&
      state.input.charCodeAt(state.position + 1) === 0x2D/* - */ &&
      state.input.charCodeAt(state.position + 2) === 0x2D/* - */) {
    state.position += 3;
    skipSeparationSpace(state, true, -1);

  } else if (hasDirectives) {
    throwError(state, 'directives end mark is expected');
  }

  composeNode(state, state.lineIndent - 1, CONTEXT_BLOCK_OUT, false, true);
  skipSeparationSpace(state, true, -1);

  if (state.checkLineBreaks &&
      PATTERN_NON_ASCII_LINE_BREAKS.test(state.input.slice(documentStart, state.position))) {
    throwWarning(state, 'non-ASCII line breaks are interpreted as content');
  }

  state.documents.push(state.result);

  if (state.position === state.lineStart && testDocumentSeparator(state)) {

    if (state.input.charCodeAt(state.position) === 0x2E/* . */) {
      state.position += 3;
      skipSeparationSpace(state, true, -1);
    }
    return;
  }

  if (state.position < (state.length - 1)) {
    throwError(state, 'end of the stream or a document separator is expected');
  } else {
    return;
  }
}


function loadDocuments(input, options) {
  input = String(input);
  options = options || {};

  if (input.length !== 0) {

    // Add tailing `\n` if not exists
    if (input.charCodeAt(input.length - 1) !== 0x0A/* LF */ &&
        input.charCodeAt(input.length - 1) !== 0x0D/* CR */) {
      input += '\n';
    }

    // Strip BOM
    if (input.charCodeAt(0) === 0xFEFF) {
      input = input.slice(1);
    }
  }

  var state = new State$1(input, options);

  var nullpos = input.indexOf('\0');

  if (nullpos !== -1) {
    state.position = nullpos;
    throwError(state, 'null byte is not allowed in input');
  }

  // Use 0 as string terminator. That significantly simplifies bounds check.
  state.input += '\0';

  while (state.input.charCodeAt(state.position) === 0x20/* Space */) {
    state.lineIndent += 1;
    state.position += 1;
  }

  while (state.position < (state.length - 1)) {
    readDocument(state);
  }

  return state.documents;
}


function loadAll$1(input, iterator, options) {
  if (iterator !== null && typeof iterator === 'object' && typeof options === 'undefined') {
    options = iterator;
    iterator = null;
  }

  var documents = loadDocuments(input, options);

  if (typeof iterator !== 'function') {
    return documents;
  }

  for (var index = 0, length = documents.length; index < length; index += 1) {
    iterator(documents[index]);
  }
}


function load$1(input, options) {
  var documents = loadDocuments(input, options);

  if (documents.length === 0) {
    /*eslint-disable no-undefined*/
    return undefined;
  } else if (documents.length === 1) {
    return documents[0];
  }
  throw new exception('expected a single document in the stream, but found more');
}


var loadAll_1 = loadAll$1;
var load_1    = load$1;

var loader = {
	loadAll: loadAll_1,
	load: load_1
};

/*eslint-disable no-use-before-define*/





var _toString       = Object.prototype.toString;
var _hasOwnProperty = Object.prototype.hasOwnProperty;

var CHAR_BOM                  = 0xFEFF;
var CHAR_TAB                  = 0x09; /* Tab */
var CHAR_LINE_FEED            = 0x0A; /* LF */
var CHAR_CARRIAGE_RETURN      = 0x0D; /* CR */
var CHAR_SPACE                = 0x20; /* Space */
var CHAR_EXCLAMATION          = 0x21; /* ! */
var CHAR_DOUBLE_QUOTE         = 0x22; /* " */
var CHAR_SHARP                = 0x23; /* # */
var CHAR_PERCENT              = 0x25; /* % */
var CHAR_AMPERSAND            = 0x26; /* & */
var CHAR_SINGLE_QUOTE         = 0x27; /* ' */
var CHAR_ASTERISK             = 0x2A; /* * */
var CHAR_COMMA                = 0x2C; /* , */
var CHAR_MINUS                = 0x2D; /* - */
var CHAR_COLON                = 0x3A; /* : */
var CHAR_EQUALS               = 0x3D; /* = */
var CHAR_GREATER_THAN         = 0x3E; /* > */
var CHAR_QUESTION             = 0x3F; /* ? */
var CHAR_COMMERCIAL_AT        = 0x40; /* @ */
var CHAR_LEFT_SQUARE_BRACKET  = 0x5B; /* [ */
var CHAR_RIGHT_SQUARE_BRACKET = 0x5D; /* ] */
var CHAR_GRAVE_ACCENT         = 0x60; /* ` */
var CHAR_LEFT_CURLY_BRACKET   = 0x7B; /* { */
var CHAR_VERTICAL_LINE        = 0x7C; /* | */
var CHAR_RIGHT_CURLY_BRACKET  = 0x7D; /* } */

var ESCAPE_SEQUENCES = {};

ESCAPE_SEQUENCES[0x00]   = '\\0';
ESCAPE_SEQUENCES[0x07]   = '\\a';
ESCAPE_SEQUENCES[0x08]   = '\\b';
ESCAPE_SEQUENCES[0x09]   = '\\t';
ESCAPE_SEQUENCES[0x0A]   = '\\n';
ESCAPE_SEQUENCES[0x0B]   = '\\v';
ESCAPE_SEQUENCES[0x0C]   = '\\f';
ESCAPE_SEQUENCES[0x0D]   = '\\r';
ESCAPE_SEQUENCES[0x1B]   = '\\e';
ESCAPE_SEQUENCES[0x22]   = '\\"';
ESCAPE_SEQUENCES[0x5C]   = '\\\\';
ESCAPE_SEQUENCES[0x85]   = '\\N';
ESCAPE_SEQUENCES[0xA0]   = '\\_';
ESCAPE_SEQUENCES[0x2028] = '\\L';
ESCAPE_SEQUENCES[0x2029] = '\\P';

var DEPRECATED_BOOLEANS_SYNTAX = [
  'y', 'Y', 'yes', 'Yes', 'YES', 'on', 'On', 'ON',
  'n', 'N', 'no', 'No', 'NO', 'off', 'Off', 'OFF'
];

var DEPRECATED_BASE60_SYNTAX = /^[-+]?[0-9_]+(?::[0-9_]+)+(?:\.[0-9_]*)?$/;

function compileStyleMap(schema, map) {
  var result, keys, index, length, tag, style, type;

  if (map === null) return {};

  result = {};
  keys = Object.keys(map);

  for (index = 0, length = keys.length; index < length; index += 1) {
    tag = keys[index];
    style = String(map[tag]);

    if (tag.slice(0, 2) === '!!') {
      tag = 'tag:yaml.org,2002:' + tag.slice(2);
    }
    type = schema.compiledTypeMap['fallback'][tag];

    if (type && _hasOwnProperty.call(type.styleAliases, style)) {
      style = type.styleAliases[style];
    }

    result[tag] = style;
  }

  return result;
}

function encodeHex(character) {
  var string, handle, length;

  string = character.toString(16).toUpperCase();

  if (character <= 0xFF) {
    handle = 'x';
    length = 2;
  } else if (character <= 0xFFFF) {
    handle = 'u';
    length = 4;
  } else if (character <= 0xFFFFFFFF) {
    handle = 'U';
    length = 8;
  } else {
    throw new exception('code point within a string may not be greater than 0xFFFFFFFF');
  }

  return '\\' + handle + common.repeat('0', length - string.length) + string;
}


var QUOTING_TYPE_SINGLE = 1,
    QUOTING_TYPE_DOUBLE = 2;

function State(options) {
  this.schema        = options['schema'] || _default;
  this.indent        = Math.max(1, (options['indent'] || 2));
  this.noArrayIndent = options['noArrayIndent'] || false;
  this.skipInvalid   = options['skipInvalid'] || false;
  this.flowLevel     = (common.isNothing(options['flowLevel']) ? -1 : options['flowLevel']);
  this.styleMap      = compileStyleMap(this.schema, options['styles'] || null);
  this.sortKeys      = options['sortKeys'] || false;
  this.lineWidth     = options['lineWidth'] || 80;
  this.noRefs        = options['noRefs'] || false;
  this.noCompatMode  = options['noCompatMode'] || false;
  this.condenseFlow  = options['condenseFlow'] || false;
  this.quotingType   = options['quotingType'] === '"' ? QUOTING_TYPE_DOUBLE : QUOTING_TYPE_SINGLE;
  this.forceQuotes   = options['forceQuotes'] || false;
  this.replacer      = typeof options['replacer'] === 'function' ? options['replacer'] : null;

  this.implicitTypes = this.schema.compiledImplicit;
  this.explicitTypes = this.schema.compiledExplicit;

  this.tag = null;
  this.result = '';

  this.duplicates = [];
  this.usedDuplicates = null;
}

// Indents every line in a string. Empty lines (\n only) are not indented.
function indentString(string, spaces) {
  var ind = common.repeat(' ', spaces),
      position = 0,
      next = -1,
      result = '',
      line,
      length = string.length;

  while (position < length) {
    next = string.indexOf('\n', position);
    if (next === -1) {
      line = string.slice(position);
      position = length;
    } else {
      line = string.slice(position, next + 1);
      position = next + 1;
    }

    if (line.length && line !== '\n') result += ind;

    result += line;
  }

  return result;
}

function generateNextLine(state, level) {
  return '\n' + common.repeat(' ', state.indent * level);
}

function testImplicitResolving(state, str) {
  var index, length, type;

  for (index = 0, length = state.implicitTypes.length; index < length; index += 1) {
    type = state.implicitTypes[index];

    if (type.resolve(str)) {
      return true;
    }
  }

  return false;
}

// [33] s-white ::= s-space | s-tab
function isWhitespace(c) {
  return c === CHAR_SPACE || c === CHAR_TAB;
}

// Returns true if the character can be printed without escaping.
// From YAML 1.2: "any allowed characters known to be non-printable
// should also be escaped. [However,] This isn’t mandatory"
// Derived from nb-char - \t - #x85 - #xA0 - #x2028 - #x2029.
function isPrintable(c) {
  return  (0x00020 <= c && c <= 0x00007E)
      || ((0x000A1 <= c && c <= 0x00D7FF) && c !== 0x2028 && c !== 0x2029)
      || ((0x0E000 <= c && c <= 0x00FFFD) && c !== CHAR_BOM)
      ||  (0x10000 <= c && c <= 0x10FFFF);
}

// [34] ns-char ::= nb-char - s-white
// [27] nb-char ::= c-printable - b-char - c-byte-order-mark
// [26] b-char  ::= b-line-feed | b-carriage-return
// Including s-white (for some reason, examples doesn't match specs in this aspect)
// ns-char ::= c-printable - b-line-feed - b-carriage-return - c-byte-order-mark
function isNsCharOrWhitespace(c) {
  return isPrintable(c)
    && c !== CHAR_BOM
    // - b-char
    && c !== CHAR_CARRIAGE_RETURN
    && c !== CHAR_LINE_FEED;
}

// [127]  ns-plain-safe(c) ::= c = flow-out  ⇒ ns-plain-safe-out
//                             c = flow-in   ⇒ ns-plain-safe-in
//                             c = block-key ⇒ ns-plain-safe-out
//                             c = flow-key  ⇒ ns-plain-safe-in
// [128] ns-plain-safe-out ::= ns-char
// [129]  ns-plain-safe-in ::= ns-char - c-flow-indicator
// [130]  ns-plain-char(c) ::=  ( ns-plain-safe(c) - “:” - “#” )
//                            | ( /* An ns-char preceding */ “#” )
//                            | ( “:” /* Followed by an ns-plain-safe(c) */ )
function isPlainSafe(c, prev, inblock) {
  var cIsNsCharOrWhitespace = isNsCharOrWhitespace(c);
  var cIsNsChar = cIsNsCharOrWhitespace && !isWhitespace(c);
  return (
    // ns-plain-safe
    inblock ? // c = flow-in
      cIsNsCharOrWhitespace
      : cIsNsCharOrWhitespace
        // - c-flow-indicator
        && c !== CHAR_COMMA
        && c !== CHAR_LEFT_SQUARE_BRACKET
        && c !== CHAR_RIGHT_SQUARE_BRACKET
        && c !== CHAR_LEFT_CURLY_BRACKET
        && c !== CHAR_RIGHT_CURLY_BRACKET
  )
    // ns-plain-char
    && c !== CHAR_SHARP // false on '#'
    && !(prev === CHAR_COLON && !cIsNsChar) // false on ': '
    || (isNsCharOrWhitespace(prev) && !isWhitespace(prev) && c === CHAR_SHARP) // change to true on '[^ ]#'
    || (prev === CHAR_COLON && cIsNsChar); // change to true on ':[^ ]'
}

// Simplified test for values allowed as the first character in plain style.
function isPlainSafeFirst(c) {
  // Uses a subset of ns-char - c-indicator
  // where ns-char = nb-char - s-white.
  // No support of ( ( “?” | “:” | “-” ) /* Followed by an ns-plain-safe(c)) */ ) part
  return isPrintable(c) && c !== CHAR_BOM
    && !isWhitespace(c) // - s-white
    // - (c-indicator ::=
    // “-” | “?” | “:” | “,” | “[” | “]” | “{” | “}”
    && c !== CHAR_MINUS
    && c !== CHAR_QUESTION
    && c !== CHAR_COLON
    && c !== CHAR_COMMA
    && c !== CHAR_LEFT_SQUARE_BRACKET
    && c !== CHAR_RIGHT_SQUARE_BRACKET
    && c !== CHAR_LEFT_CURLY_BRACKET
    && c !== CHAR_RIGHT_CURLY_BRACKET
    // | “#” | “&” | “*” | “!” | “|” | “=” | “>” | “'” | “"”
    && c !== CHAR_SHARP
    && c !== CHAR_AMPERSAND
    && c !== CHAR_ASTERISK
    && c !== CHAR_EXCLAMATION
    && c !== CHAR_VERTICAL_LINE
    && c !== CHAR_EQUALS
    && c !== CHAR_GREATER_THAN
    && c !== CHAR_SINGLE_QUOTE
    && c !== CHAR_DOUBLE_QUOTE
    // | “%” | “@” | “`”)
    && c !== CHAR_PERCENT
    && c !== CHAR_COMMERCIAL_AT
    && c !== CHAR_GRAVE_ACCENT;
}

// Simplified test for values allowed as the last character in plain style.
function isPlainSafeLast(c) {
  // just not whitespace or colon, it will be checked to be plain character later
  return !isWhitespace(c) && c !== CHAR_COLON;
}

// Same as 'string'.codePointAt(pos), but works in older browsers.
function codePointAt(string, pos) {
  var first = string.charCodeAt(pos), second;
  if (first >= 0xD800 && first <= 0xDBFF && pos + 1 < string.length) {
    second = string.charCodeAt(pos + 1);
    if (second >= 0xDC00 && second <= 0xDFFF) {
      // https://mathiasbynens.be/notes/javascript-encoding#surrogate-formulae
      return (first - 0xD800) * 0x400 + second - 0xDC00 + 0x10000;
    }
  }
  return first;
}

// Determines whether block indentation indicator is required.
function needIndentIndicator(string) {
  var leadingSpaceRe = /^\n* /;
  return leadingSpaceRe.test(string);
}

var STYLE_PLAIN   = 1,
    STYLE_SINGLE  = 2,
    STYLE_LITERAL = 3,
    STYLE_FOLDED  = 4,
    STYLE_DOUBLE  = 5;

// Determines which scalar styles are possible and returns the preferred style.
// lineWidth = -1 => no limit.
// Pre-conditions: str.length > 0.
// Post-conditions:
//    STYLE_PLAIN or STYLE_SINGLE => no \n are in the string.
//    STYLE_LITERAL => no lines are suitable for folding (or lineWidth is -1).
//    STYLE_FOLDED => a line > lineWidth and can be folded (and lineWidth != -1).
function chooseScalarStyle(string, singleLineOnly, indentPerLevel, lineWidth,
  testAmbiguousType, quotingType, forceQuotes, inblock) {

  var i;
  var char = 0;
  var prevChar = null;
  var hasLineBreak = false;
  var hasFoldableLine = false; // only checked if shouldTrackWidth
  var shouldTrackWidth = lineWidth !== -1;
  var previousLineBreak = -1; // count the first line correctly
  var plain = isPlainSafeFirst(codePointAt(string, 0))
          && isPlainSafeLast(codePointAt(string, string.length - 1));

  if (singleLineOnly || forceQuotes) {
    // Case: no block styles.
    // Check for disallowed characters to rule out plain and single.
    for (i = 0; i < string.length; char >= 0x10000 ? i += 2 : i++) {
      char = codePointAt(string, i);
      if (!isPrintable(char)) {
        return STYLE_DOUBLE;
      }
      plain = plain && isPlainSafe(char, prevChar, inblock);
      prevChar = char;
    }
  } else {
    // Case: block styles permitted.
    for (i = 0; i < string.length; char >= 0x10000 ? i += 2 : i++) {
      char = codePointAt(string, i);
      if (char === CHAR_LINE_FEED) {
        hasLineBreak = true;
        // Check if any line can be folded.
        if (shouldTrackWidth) {
          hasFoldableLine = hasFoldableLine ||
            // Foldable line = too long, and not more-indented.
            (i - previousLineBreak - 1 > lineWidth &&
             string[previousLineBreak + 1] !== ' ');
          previousLineBreak = i;
        }
      } else if (!isPrintable(char)) {
        return STYLE_DOUBLE;
      }
      plain = plain && isPlainSafe(char, prevChar, inblock);
      prevChar = char;
    }
    // in case the end is missing a \n
    hasFoldableLine = hasFoldableLine || (shouldTrackWidth &&
      (i - previousLineBreak - 1 > lineWidth &&
       string[previousLineBreak + 1] !== ' '));
  }
  // Although every style can represent \n without escaping, prefer block styles
  // for multiline, since they're more readable and they don't add empty lines.
  // Also prefer folding a super-long line.
  if (!hasLineBreak && !hasFoldableLine) {
    // Strings interpretable as another type have to be quoted;
    // e.g. the string 'true' vs. the boolean true.
    if (plain && !forceQuotes && !testAmbiguousType(string)) {
      return STYLE_PLAIN;
    }
    return quotingType === QUOTING_TYPE_DOUBLE ? STYLE_DOUBLE : STYLE_SINGLE;
  }
  // Edge case: block indentation indicator can only have one digit.
  if (indentPerLevel > 9 && needIndentIndicator(string)) {
    return STYLE_DOUBLE;
  }
  // At this point we know block styles are valid.
  // Prefer literal style unless we want to fold.
  if (!forceQuotes) {
    return hasFoldableLine ? STYLE_FOLDED : STYLE_LITERAL;
  }
  return quotingType === QUOTING_TYPE_DOUBLE ? STYLE_DOUBLE : STYLE_SINGLE;
}

// Note: line breaking/folding is implemented for only the folded style.
// NB. We drop the last trailing newline (if any) of a returned block scalar
//  since the dumper adds its own newline. This always works:
//    • No ending newline => unaffected; already using strip "-" chomping.
//    • Ending newline    => removed then restored.
//  Importantly, this keeps the "+" chomp indicator from gaining an extra line.
function writeScalar(state, string, level, iskey, inblock) {
  state.dump = (function () {
    if (string.length === 0) {
      return state.quotingType === QUOTING_TYPE_DOUBLE ? '""' : "''";
    }
    if (!state.noCompatMode) {
      if (DEPRECATED_BOOLEANS_SYNTAX.indexOf(string) !== -1 || DEPRECATED_BASE60_SYNTAX.test(string)) {
        return state.quotingType === QUOTING_TYPE_DOUBLE ? ('"' + string + '"') : ("'" + string + "'");
      }
    }

    var indent = state.indent * Math.max(1, level); // no 0-indent scalars
    // As indentation gets deeper, let the width decrease monotonically
    // to the lower bound min(state.lineWidth, 40).
    // Note that this implies
    //  state.lineWidth ≤ 40 + state.indent: width is fixed at the lower bound.
    //  state.lineWidth > 40 + state.indent: width decreases until the lower bound.
    // This behaves better than a constant minimum width which disallows narrower options,
    // or an indent threshold which causes the width to suddenly increase.
    var lineWidth = state.lineWidth === -1
      ? -1 : Math.max(Math.min(state.lineWidth, 40), state.lineWidth - indent);

    // Without knowing if keys are implicit/explicit, assume implicit for safety.
    var singleLineOnly = iskey
      // No block styles in flow mode.
      || (state.flowLevel > -1 && level >= state.flowLevel);
    function testAmbiguity(string) {
      return testImplicitResolving(state, string);
    }

    switch (chooseScalarStyle(string, singleLineOnly, state.indent, lineWidth,
      testAmbiguity, state.quotingType, state.forceQuotes && !iskey, inblock)) {

      case STYLE_PLAIN:
        return string;
      case STYLE_SINGLE:
        return "'" + string.replace(/'/g, "''") + "'";
      case STYLE_LITERAL:
        return '|' + blockHeader(string, state.indent)
          + dropEndingNewline(indentString(string, indent));
      case STYLE_FOLDED:
        return '>' + blockHeader(string, state.indent)
          + dropEndingNewline(indentString(foldString(string, lineWidth), indent));
      case STYLE_DOUBLE:
        return '"' + escapeString(string) + '"';
      default:
        throw new exception('impossible error: invalid scalar style');
    }
  }());
}

// Pre-conditions: string is valid for a block scalar, 1 <= indentPerLevel <= 9.
function blockHeader(string, indentPerLevel) {
  var indentIndicator = needIndentIndicator(string) ? String(indentPerLevel) : '';

  // note the special case: the string '\n' counts as a "trailing" empty line.
  var clip =          string[string.length - 1] === '\n';
  var keep = clip && (string[string.length - 2] === '\n' || string === '\n');
  var chomp = keep ? '+' : (clip ? '' : '-');

  return indentIndicator + chomp + '\n';
}

// (See the note for writeScalar.)
function dropEndingNewline(string) {
  return string[string.length - 1] === '\n' ? string.slice(0, -1) : string;
}

// Note: a long line without a suitable break point will exceed the width limit.
// Pre-conditions: every char in str isPrintable, str.length > 0, width > 0.
function foldString(string, width) {
  // In folded style, $k$ consecutive newlines output as $k+1$ newlines—
  // unless they're before or after a more-indented line, or at the very
  // beginning or end, in which case $k$ maps to $k$.
  // Therefore, parse each chunk as newline(s) followed by a content line.
  var lineRe = /(\n+)([^\n]*)/g;

  // first line (possibly an empty line)
  var result = (function () {
    var nextLF = string.indexOf('\n');
    nextLF = nextLF !== -1 ? nextLF : string.length;
    lineRe.lastIndex = nextLF;
    return foldLine(string.slice(0, nextLF), width);
  }());
  // If we haven't reached the first content line yet, don't add an extra \n.
  var prevMoreIndented = string[0] === '\n' || string[0] === ' ';
  var moreIndented;

  // rest of the lines
  var match;
  while ((match = lineRe.exec(string))) {
    var prefix = match[1], line = match[2];
    moreIndented = (line[0] === ' ');
    result += prefix
      + (!prevMoreIndented && !moreIndented && line !== ''
        ? '\n' : '')
      + foldLine(line, width);
    prevMoreIndented = moreIndented;
  }

  return result;
}

// Greedy line breaking.
// Picks the longest line under the limit each time,
// otherwise settles for the shortest line over the limit.
// NB. More-indented lines *cannot* be folded, as that would add an extra \n.
function foldLine(line, width) {
  if (line === '' || line[0] === ' ') return line;

  // Since a more-indented line adds a \n, breaks can't be followed by a space.
  var breakRe = / [^ ]/g; // note: the match index will always be <= length-2.
  var match;
  // start is an inclusive index. end, curr, and next are exclusive.
  var start = 0, end, curr = 0, next = 0;
  var result = '';

  // Invariants: 0 <= start <= length-1.
  //   0 <= curr <= next <= max(0, length-2). curr - start <= width.
  // Inside the loop:
  //   A match implies length >= 2, so curr and next are <= length-2.
  while ((match = breakRe.exec(line))) {
    next = match.index;
    // maintain invariant: curr - start <= width
    if (next - start > width) {
      end = (curr > start) ? curr : next; // derive end <= length-2
      result += '\n' + line.slice(start, end);
      // skip the space that was output as \n
      start = end + 1;                    // derive start <= length-1
    }
    curr = next;
  }

  // By the invariants, start <= length-1, so there is something left over.
  // It is either the whole string or a part starting from non-whitespace.
  result += '\n';
  // Insert a break if the remainder is too long and there is a break available.
  if (line.length - start > width && curr > start) {
    result += line.slice(start, curr) + '\n' + line.slice(curr + 1);
  } else {
    result += line.slice(start);
  }

  return result.slice(1); // drop extra \n joiner
}

// Escapes a double-quoted string.
function escapeString(string) {
  var result = '';
  var char = 0;
  var escapeSeq;

  for (var i = 0; i < string.length; char >= 0x10000 ? i += 2 : i++) {
    char = codePointAt(string, i);
    escapeSeq = ESCAPE_SEQUENCES[char];

    if (!escapeSeq && isPrintable(char)) {
      result += string[i];
      if (char >= 0x10000) result += string[i + 1];
    } else {
      result += escapeSeq || encodeHex(char);
    }
  }

  return result;
}

function writeFlowSequence(state, level, object) {
  var _result = '',
      _tag    = state.tag,
      index,
      length,
      value;

  for (index = 0, length = object.length; index < length; index += 1) {
    value = object[index];

    if (state.replacer) {
      value = state.replacer.call(object, String(index), value);
    }

    // Write only valid elements, put null instead of invalid elements.
    if (writeNode(state, level, value, false, false) ||
        (typeof value === 'undefined' &&
         writeNode(state, level, null, false, false))) {

      if (_result !== '') _result += ',' + (!state.condenseFlow ? ' ' : '');
      _result += state.dump;
    }
  }

  state.tag = _tag;
  state.dump = '[' + _result + ']';
}

function writeBlockSequence(state, level, object, compact) {
  var _result = '',
      _tag    = state.tag,
      index,
      length,
      value;

  for (index = 0, length = object.length; index < length; index += 1) {
    value = object[index];

    if (state.replacer) {
      value = state.replacer.call(object, String(index), value);
    }

    // Write only valid elements, put null instead of invalid elements.
    if (writeNode(state, level + 1, value, true, true, false, true) ||
        (typeof value === 'undefined' &&
         writeNode(state, level + 1, null, true, true, false, true))) {

      if (!compact || _result !== '') {
        _result += generateNextLine(state, level);
      }

      if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
        _result += '-';
      } else {
        _result += '- ';
      }

      _result += state.dump;
    }
  }

  state.tag = _tag;
  state.dump = _result || '[]'; // Empty sequence if no valid values.
}

function writeFlowMapping(state, level, object) {
  var _result       = '',
      _tag          = state.tag,
      objectKeyList = Object.keys(object),
      index,
      length,
      objectKey,
      objectValue,
      pairBuffer;

  for (index = 0, length = objectKeyList.length; index < length; index += 1) {

    pairBuffer = '';
    if (_result !== '') pairBuffer += ', ';

    if (state.condenseFlow) pairBuffer += '"';

    objectKey = objectKeyList[index];
    objectValue = object[objectKey];

    if (state.replacer) {
      objectValue = state.replacer.call(object, objectKey, objectValue);
    }

    if (!writeNode(state, level, objectKey, false, false)) {
      continue; // Skip this pair because of invalid key;
    }

    if (state.dump.length > 1024) pairBuffer += '? ';

    pairBuffer += state.dump + (state.condenseFlow ? '"' : '') + ':' + (state.condenseFlow ? '' : ' ');

    if (!writeNode(state, level, objectValue, false, false)) {
      continue; // Skip this pair because of invalid value.
    }

    pairBuffer += state.dump;

    // Both key and value are valid.
    _result += pairBuffer;
  }

  state.tag = _tag;
  state.dump = '{' + _result + '}';
}

function writeBlockMapping(state, level, object, compact) {
  var _result       = '',
      _tag          = state.tag,
      objectKeyList = Object.keys(object),
      index,
      length,
      objectKey,
      objectValue,
      explicitPair,
      pairBuffer;

  // Allow sorting keys so that the output file is deterministic
  if (state.sortKeys === true) {
    // Default sorting
    objectKeyList.sort();
  } else if (typeof state.sortKeys === 'function') {
    // Custom sort function
    objectKeyList.sort(state.sortKeys);
  } else if (state.sortKeys) {
    // Something is wrong
    throw new exception('sortKeys must be a boolean or a function');
  }

  for (index = 0, length = objectKeyList.length; index < length; index += 1) {
    pairBuffer = '';

    if (!compact || _result !== '') {
      pairBuffer += generateNextLine(state, level);
    }

    objectKey = objectKeyList[index];
    objectValue = object[objectKey];

    if (state.replacer) {
      objectValue = state.replacer.call(object, objectKey, objectValue);
    }

    if (!writeNode(state, level + 1, objectKey, true, true, true)) {
      continue; // Skip this pair because of invalid key.
    }

    explicitPair = (state.tag !== null && state.tag !== '?') ||
                   (state.dump && state.dump.length > 1024);

    if (explicitPair) {
      if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
        pairBuffer += '?';
      } else {
        pairBuffer += '? ';
      }
    }

    pairBuffer += state.dump;

    if (explicitPair) {
      pairBuffer += generateNextLine(state, level);
    }

    if (!writeNode(state, level + 1, objectValue, true, explicitPair)) {
      continue; // Skip this pair because of invalid value.
    }

    if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
      pairBuffer += ':';
    } else {
      pairBuffer += ': ';
    }

    pairBuffer += state.dump;

    // Both key and value are valid.
    _result += pairBuffer;
  }

  state.tag = _tag;
  state.dump = _result || '{}'; // Empty mapping if no valid pairs.
}

function detectType(state, object, explicit) {
  var _result, typeList, index, length, type, style;

  typeList = explicit ? state.explicitTypes : state.implicitTypes;

  for (index = 0, length = typeList.length; index < length; index += 1) {
    type = typeList[index];

    if ((type.instanceOf  || type.predicate) &&
        (!type.instanceOf || ((typeof object === 'object') && (object instanceof type.instanceOf))) &&
        (!type.predicate  || type.predicate(object))) {

      if (explicit) {
        if (type.multi && type.representName) {
          state.tag = type.representName(object);
        } else {
          state.tag = type.tag;
        }
      } else {
        state.tag = '?';
      }

      if (type.represent) {
        style = state.styleMap[type.tag] || type.defaultStyle;

        if (_toString.call(type.represent) === '[object Function]') {
          _result = type.represent(object, style);
        } else if (_hasOwnProperty.call(type.represent, style)) {
          _result = type.represent[style](object, style);
        } else {
          throw new exception('!<' + type.tag + '> tag resolver accepts not "' + style + '" style');
        }

        state.dump = _result;
      }

      return true;
    }
  }

  return false;
}

// Serializes `object` and writes it to global `result`.
// Returns true on success, or false on invalid object.
//
function writeNode(state, level, object, block, compact, iskey, isblockseq) {
  state.tag = null;
  state.dump = object;

  if (!detectType(state, object, false)) {
    detectType(state, object, true);
  }

  var type = _toString.call(state.dump);
  var inblock = block;
  var tagStr;

  if (block) {
    block = (state.flowLevel < 0 || state.flowLevel > level);
  }

  var objectOrArray = type === '[object Object]' || type === '[object Array]',
      duplicateIndex,
      duplicate;

  if (objectOrArray) {
    duplicateIndex = state.duplicates.indexOf(object);
    duplicate = duplicateIndex !== -1;
  }

  if ((state.tag !== null && state.tag !== '?') || duplicate || (state.indent !== 2 && level > 0)) {
    compact = false;
  }

  if (duplicate && state.usedDuplicates[duplicateIndex]) {
    state.dump = '*ref_' + duplicateIndex;
  } else {
    if (objectOrArray && duplicate && !state.usedDuplicates[duplicateIndex]) {
      state.usedDuplicates[duplicateIndex] = true;
    }
    if (type === '[object Object]') {
      if (block && (Object.keys(state.dump).length !== 0)) {
        writeBlockMapping(state, level, state.dump, compact);
        if (duplicate) {
          state.dump = '&ref_' + duplicateIndex + state.dump;
        }
      } else {
        writeFlowMapping(state, level, state.dump);
        if (duplicate) {
          state.dump = '&ref_' + duplicateIndex + ' ' + state.dump;
        }
      }
    } else if (type === '[object Array]') {
      if (block && (state.dump.length !== 0)) {
        if (state.noArrayIndent && !isblockseq && level > 0) {
          writeBlockSequence(state, level - 1, state.dump, compact);
        } else {
          writeBlockSequence(state, level, state.dump, compact);
        }
        if (duplicate) {
          state.dump = '&ref_' + duplicateIndex + state.dump;
        }
      } else {
        writeFlowSequence(state, level, state.dump);
        if (duplicate) {
          state.dump = '&ref_' + duplicateIndex + ' ' + state.dump;
        }
      }
    } else if (type === '[object String]') {
      if (state.tag !== '?') {
        writeScalar(state, state.dump, level, iskey, inblock);
      }
    } else if (type === '[object Undefined]') {
      return false;
    } else {
      if (state.skipInvalid) return false;
      throw new exception('unacceptable kind of an object to dump ' + type);
    }

    if (state.tag !== null && state.tag !== '?') {
      // Need to encode all characters except those allowed by the spec:
      //
      // [35] ns-dec-digit    ::=  [#x30-#x39] /* 0-9 */
      // [36] ns-hex-digit    ::=  ns-dec-digit
      //                         | [#x41-#x46] /* A-F */ | [#x61-#x66] /* a-f */
      // [37] ns-ascii-letter ::=  [#x41-#x5A] /* A-Z */ | [#x61-#x7A] /* a-z */
      // [38] ns-word-char    ::=  ns-dec-digit | ns-ascii-letter | “-”
      // [39] ns-uri-char     ::=  “%” ns-hex-digit ns-hex-digit | ns-word-char | “#”
      //                         | “;” | “/” | “?” | “:” | “@” | “&” | “=” | “+” | “$” | “,”
      //                         | “_” | “.” | “!” | “~” | “*” | “'” | “(” | “)” | “[” | “]”
      //
      // Also need to encode '!' because it has special meaning (end of tag prefix).
      //
      tagStr = encodeURI(
        state.tag[0] === '!' ? state.tag.slice(1) : state.tag
      ).replace(/!/g, '%21');

      if (state.tag[0] === '!') {
        tagStr = '!' + tagStr;
      } else if (tagStr.slice(0, 18) === 'tag:yaml.org,2002:') {
        tagStr = '!!' + tagStr.slice(18);
      } else {
        tagStr = '!<' + tagStr + '>';
      }

      state.dump = tagStr + ' ' + state.dump;
    }
  }

  return true;
}

function getDuplicateReferences(object, state) {
  var objects = [],
      duplicatesIndexes = [],
      index,
      length;

  inspectNode(object, objects, duplicatesIndexes);

  for (index = 0, length = duplicatesIndexes.length; index < length; index += 1) {
    state.duplicates.push(objects[duplicatesIndexes[index]]);
  }
  state.usedDuplicates = new Array(length);
}

function inspectNode(object, objects, duplicatesIndexes) {
  var objectKeyList,
      index,
      length;

  if (object !== null && typeof object === 'object') {
    index = objects.indexOf(object);
    if (index !== -1) {
      if (duplicatesIndexes.indexOf(index) === -1) {
        duplicatesIndexes.push(index);
      }
    } else {
      objects.push(object);

      if (Array.isArray(object)) {
        for (index = 0, length = object.length; index < length; index += 1) {
          inspectNode(object[index], objects, duplicatesIndexes);
        }
      } else {
        objectKeyList = Object.keys(object);

        for (index = 0, length = objectKeyList.length; index < length; index += 1) {
          inspectNode(object[objectKeyList[index]], objects, duplicatesIndexes);
        }
      }
    }
  }
}

function dump$1(input, options) {
  options = options || {};

  var state = new State(options);

  if (!state.noRefs) getDuplicateReferences(input, state);

  var value = input;

  if (state.replacer) {
    value = state.replacer.call({ '': value }, '', value);
  }

  if (writeNode(state, 0, value, true, true)) return state.dump + '\n';

  return '';
}

var dump_1 = dump$1;

var dumper = {
	dump: dump_1
};

function renamed(from, to) {
  return function () {
    throw new Error('Function yaml.' + from + ' is removed in js-yaml 4. ' +
      'Use yaml.' + to + ' instead, which is now safe by default.');
  };
}


var Type                = type;
var Schema              = schema;
var FAILSAFE_SCHEMA     = failsafe;
var JSON_SCHEMA         = json;
var CORE_SCHEMA         = core;
var DEFAULT_SCHEMA      = _default;
var load                = loader.load;
var loadAll             = loader.loadAll;
var dump                = dumper.dump;
var YAMLException       = exception;

// Re-export all types in case user wants to create custom schema
var types = {
  binary:    binary,
  float:     float,
  map:       map,
  null:      _null,
  pairs:     pairs,
  set:       set,
  timestamp: timestamp,
  bool:      bool,
  int:       int,
  merge:     merge,
  omap:      omap,
  seq:       seq,
  str:       str
};

// Removed functions from JS-YAML 3.0.x
var safeLoad            = renamed('safeLoad', 'load');
var safeLoadAll         = renamed('safeLoadAll', 'loadAll');
var safeDump            = renamed('safeDump', 'dump');

var jsYaml = {
	Type: Type,
	Schema: Schema,
	FAILSAFE_SCHEMA: FAILSAFE_SCHEMA,
	JSON_SCHEMA: JSON_SCHEMA,
	CORE_SCHEMA: CORE_SCHEMA,
	DEFAULT_SCHEMA: DEFAULT_SCHEMA,
	load: load,
	loadAll: loadAll,
	dump: dump,
	YAMLException: YAMLException,
	types: types,
	safeLoad: safeLoad,
	safeLoadAll: safeLoadAll,
	safeDump: safeDump
};

function getModuleFromYaml(moduleName, yamlPath) {
    const data = jsYaml.loadAll(fs.readFileSync(yamlPath, 'utf8'));
    const mod = data[0].modules[moduleName];
    if (mod) {
        // Convert fields from object to array
        mod.fields = Object.keys(mod.fields).map((k) => mod.fields[k]);
        return new Module(mod);
    }
    else {
        return undefined;
    }
}

// The default dataset post processing function to use.
// This one simply returns the current value.
const defaultFx = 'n';
/**
 * BaseChart represents a structure that stores any configuration data.
 * Any display and data rendering operations should be handled by any sub classes.
 */
class BaseChart {
    constructor(def = {}) {
        this.chartID = NoID;
        this.namespaceID = NoID;
        this.name = '';
        this.handle = '';
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.canUpdateChart = false;
        this.canDeleteChart = false;
        this.canGrant = false;
        this.config = {};
        this.merge(def);
    }
    /**
     * The method performs post processing for each value in the given dataset.
     * It works with a simple equation written in javascript (example: n + m).
     * Available variables to use:
     * * n - current value
     * * m - previous value (undefined in case of the first element)
     * * r - entire data array.
     *
     * @param data Array of values in the given data set
     * @param m Metric for the given dataset
     */
    datasetPostProc(data, m) {
        // Define a valid function to evaluate
        let fxRaw = (m.fx || defaultFx).trim();
        if (!fxRaw.startsWith('return')) {
            fxRaw = 'return ' + fxRaw;
        }
        // eslint-disable-next-line no-new-func
        const fx = new Function('n', 'm', 'r', fxRaw);
        // Define a new array, so we don't alter the original one.
        const r = [...data];
        // Run postprocessing for all data in the given data set
        // There is a slight difference between temporal data points and categorical data points.
        if (data[0] instanceof Object) {
            // Temporal
            for (let i = 0; i < data.length; i++) {
                const a = data[i];
                const b = data[i - 1];
                const n = a.y;
                let m;
                if (i > 0) {
                    m = b === null || b === void 0 ? void 0 : b.y;
                }
                a.y = fx(n, m, r);
            }
        }
        else {
            // Categorical
            for (let i = 0; i < data.length; i++) {
                const n = data[i];
                let m;
                if (i > 0) {
                    m = data[i - 1];
                }
                data[i] = fx(n, m, r);
            }
        }
        return data;
    }
    merge(c) {
        var _a;
        let conf = Object.assign({}, (c.config || {}));
        Apply(this, c, CortezaID, 'chartID', 'namespaceID');
        Apply(this, c, String, 'name', 'handle');
        Apply(this, c, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, c, Boolean, 'canUpdateChart', 'canDeleteChart', 'canGrant');
        Apply(this, c, Object, 'config');
        if (typeof c.config === 'object') {
            // Verify & normalize
            const _b = c.config, { reports = [] } = _b, rest = __rest(_b, ["reports"]);
            conf = Object.assign({ reports: reports || [] }, rest);
        }
        this.config = (conf ? lodash.merge(this.defConfig(), conf) : false) || this.config || this.defConfig();
        (_a = this.config.reports) === null || _a === void 0 ? void 0 : _a.forEach(report => {
            const { dimensions = [], metrics = [] } = report || {};
            report.dimensions = dimensions.map(d => {
                // Legacy support
                if (d.modifier === 'auto') {
                    d.timeLabels = true;
                    d.modifier = '(no grouping / buckets)';
                }
                if (d.field === 'created_at') {
                    d.field = 'createdAt';
                }
                return lodash.merge(this.defDimension(), d);
            });
            report.metrics = metrics.map(m => lodash.merge(this.defMetric(), m));
        });
    }
    /**
     * Checks reports validity.
     * Validates dimensions and metrics.
     * If invalid it throws an error.
     */
    isValid() {
        if (!this.config.reports || !this.config.reports.length) {
            throw new Error('notification.chart.invalidConfig.missingReports');
        }
        this.config.reports.forEach(({ moduleID, dimensions, metrics }) => {
            if (!moduleID) {
                throw new Error('notification.chart.invalidConfig.missingModuleID');
            }
            // Expecting all dimensions to have defined fields
            dimensions === null || dimensions === void 0 ? void 0 : dimensions.forEach(this.dimCheck);
            // Expecting all metrics to have defined fields
            metrics === null || metrics === void 0 ? void 0 : metrics.forEach(this.mtrCheck);
        });
        return true;
    }
    /**
     * Checks validity of dimensions.
     * If invalid it throws an error
     */
    dimCheck({ field, modifier }) {
        if (!field) {
            throw new Error('notification.chart.invalidConfig.missingDimensionsField');
        }
        if (!modifier) {
            throw new Error('notification.chart.invalidConfig.missingDimensionsModifier');
        }
    }
    /**
     * Checks validity of metrics.
     * If invalid it throws an error
     */
    mtrCheck({ field, aggregate, type }) {
        if (!field) {
            throw new Error('notification.chart.invalidConfig.missingMetricsField');
        }
        if (field !== 'count' && !aggregate) {
            throw new Error('notification.chart.invalidConfig.missingMetricsAggregate');
        }
        if (!type) {
            throw new Error('notification.chart.invalidConfig.missingMetricsType');
        }
    }
    /**
     * Prepares params that the reporter can use for querying.
     */
    formatReporterParams({ moduleID, metrics, dimensions, filter }) {
        return {
            moduleID,
            filter,
            // Remove count (we'll get it anyway) and construct FUNC(ARG) params
            metrics: metrics === null || metrics === void 0 ? void 0 : metrics.filter((m) => m.field !== 'count').map((m) => `${m.aggregate}(${m.field}) AS ${makeAlias(m)}`).join(','),
            // Construct dimensions \w modifiers...
            dimensions: dimensions === null || dimensions === void 0 ? void 0 : dimensions.map(d => (Object.assign({ field: 'createdAt' }, d))).map((d) => dimensionFunctions.convert(d))[0],
        };
    }
    /**
     * Fetcher reports defined in the given configuration with the help of the provided
     * reporter.
     */
    fetchReports(_a) {
        return __awaiter(this, arguments, void 0, function* ({ reporter }) {
            var _b;
            const out = [];
            // Prepare params & filter out invalid combos (formatReporterParams will return null on invalid params)
            const reports = (_b = this.config.reports) === null || _b === void 0 ? void 0 : _b.map(this.formatReporterParams).map(params => reporter(params)).map((p, index) => p.then((results) => {
                results = results || [];
                out[index] = this.processReporterResults(results, (this.config.reports || [])[index]);
            }));
            // Wait for all requests to finish and return new promise, with results
            return Promise.all(reports).then(() => new Promise(resolve => {
                resolve(out);
            }));
        });
    }
    /**
     * Processes provided report with it's results:
     * * skip missing values, if so requested,
     * * generate labels,
     * * creates dataset for the chart.
     */
    processReporterResults(results = [], report) {
        var _a;
        const dLabel = 'dimension_0';
        const { dimensions: [dimension] = [] } = report;
        let labels = [];
        // helper to choose between eight the provided value, default value or a generic 'undefined'
        const pickValue = (val, { default: dDft }) => {
            return val || val === 0 ? val : dDft || 'undefined';
        };
        // Skip missing values; if so requested
        if (dimension.skipMissing) {
            results = results.filter((r) => r[dLabel] || r[dLabel] === 0);
        }
        // Not a time dimensions, build set of labels
        labels = results.map((r) => pickValue(r[dLabel], dimension));
        // Build data sets
        const datasets = (_a = report.metrics) === null || _a === void 0 ? void 0 : _a.map(m => {
            const alias = makeAlias({ field: m.field, aggregate: m.aggregate });
            const data = results.map((r) => {
                return pickValue(r[m.field === 'count' ? m.field : alias], dimension);
            });
            // Any sub class has the ability to define how the dataset looks like.
            // this comes in handy when we want to support charts with different definitions.
            return this.makeDataset(m, dimension, data, alias);
        });
        return {
            labels: this.processLabels(labels, dimension),
            datasets,
            dimension,
        };
    }
    processLabels(ll, d) {
        return ll;
    }
    makeDataset(m, d, data, alias) {
        throw new Error('method.makeDataset.notImplemented');
    }
    makeOptions(data) {
        throw new Error('method.makeOptions.notImplemented');
    }
    plugins(mm) {
        throw new Error('method.plugins.notImplemented');
    }
    baseChartType(datasets) {
        throw new Error('method.baseChartType.notImplemented');
    }
    /**
     * Performs chart export; used by exporter feature.
     */
    export(findModuleByID) {
        return __awaiter(this, void 0, void 0, function* () {
            var _a;
            const { namespaceID } = this;
            const copy = new BaseChart(this);
            if ((_a = copy.config) === null || _a === void 0 ? void 0 : _a.reports) {
                yield Promise.all(copy.config.reports.map((r) => __awaiter(this, void 0, void 0, function* () {
                    const { moduleID } = r;
                    if (moduleID) {
                        const module = yield findModuleByID({ namespaceID, moduleID });
                        r.moduleID = module.name;
                        return r;
                    }
                    else {
                        return null;
                    }
                }))).then((a) => {
                    return a;
                });
            }
            return copy;
        });
    }
    /**
     * Performs import; used by importer feature
     */
    import(getModuleID) {
        var _a, _b;
        const copy = new BaseChart(this);
        copy.config.reports = (_b = (_a = copy.config) === null || _a === void 0 ? void 0 : _a.reports) === null || _b === void 0 ? void 0 : _b.map(r => {
            const { moduleID } = r;
            if (moduleID) {
                r.moduleID = getModuleID(moduleID);
            }
            return r;
        });
        return copy;
    }
    defDimension() {
        return Object.assign({}, {
            conditions: {},
            meta: {},
            rotateLabel: 0,
        });
    }
    defMetric() {
        return Object.assign({}, {
            formatting: defFormatData(),
        });
    }
    defReport() {
        return Object.assign({}, {
            moduleID: undefined,
            filter: '',
            dimensions: [this.defDimension()],
            metrics: [this.defMetric()],
            yAxis: {
                axisType: 'linear',
                axisPosition: 'left',
                labelPosition: 'end',
                rotateLabel: 0,
                formatting: defFormatData(),
            },
            tooltip: {},
            legend: {
                isScrollable: true,
                orientation: 'horizontal',
                align: 'center',
                position: {
                    top: undefined,
                    right: undefined,
                    bottom: undefined,
                    left: undefined,
                    isDefault: true,
                },
            },
            offset: {
                top: '50',
                right: '30',
                bottom: '20',
                left: '30',
                isDefault: true,
            },
        });
    }
    defConfig() {
        return Object.assign({}, {
            colorScheme: '',
            reports: [this.defReport()],
            noAnimation: false,
            toolbox: {
                saveAsImage: false,
                timeline: '',
            },
        });
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'compose:chart';
    }
    clone() {
        return new BaseChart(JSON.parse(JSON.stringify(this)));
    }
}

/**
 * Chart represents a generic chart, such as a bar chart, line chart, ...
 */
class Chart extends BaseChart {
    // Generic charts (at the moment) support only 1 report per chart
    fetchReports(a) {
        const _super = Object.create(null, {
            fetchReports: { get: () => super.fetchReports }
        });
        return __awaiter(this, void 0, void 0, function* () {
            return _super.fetchReports.call(this, a).then((rr) => {
                return rr[0];
            });
        });
    }
    makeDataset(m, d, data, alias) {
        data = this.datasetPostProc(data, m);
        return {
            type: m.type,
            label: m.label || m.field,
            data,
            fill: m.fill,
            smooth: m.smooth,
            step: m.step ? 'middle' : undefined,
            roseType: m.rose ? 'radius' : undefined,
            symbol: m.symbol,
            stack: m.stack,
            tooltip: {
                fixed: m.fixTooltips,
                relative: m.relativeValue && !['bar', 'line'].includes(m.type),
            },
            formatting: m.formatting,
        };
    }
    makeOptions(data) {
        var _a, _b, _c, _d, _e, _f, _g, _h;
        const { reports = [], colorScheme, noAnimation = false, toolbox } = this.config;
        const { saveAsImage, timeline = '' } = toolbox || {};
        const options = {
            animation: !noAnimation,
            series: [],
            xAxis: [],
            yAxis: [],
            tooltip: {
                show: true,
                appendToBody: true,
                position: 'inside',
            },
        };
        const { labels, datasets = [], themeVariables = {} } = data;
        const { dimensions: [dimension] = [], yAxis, offset, tooltip: t, legend: l, } = reports[0] || {};
        const hasAxis = datasets.some(({ type }) => ['bar', 'line', 'scatter'].includes(type));
        let horizontal = false;
        if (hasAxis) {
            if (yAxis) {
                const { label: yLabel, axisType: yType = 'linear', axisPosition: position = 'left', labelPosition = 'end', beginAtZero, min, max, } = yAxis;
                horizontal = !!yAxis.horizontal;
                const xAxis = {
                    nameLocation: 'center',
                    type: dimension.timeLabels ? 'time' : 'category',
                    axisLabel: {
                        interval: 0,
                        overflow: 'break',
                        hideOverlap: true,
                        rotate: dimension.rotateLabel,
                    },
                    axisTick: {
                        show: false,
                    },
                    axisLine: {
                        show: false,
                    },
                };
                const tempYAxis = {
                    name: yLabel,
                    type: yType === 'linear' ? 'value' : 'log',
                    position,
                    nameLocation: labelPosition,
                    min: beginAtZero ? 0 : Number(min) || undefined,
                    max: Number(max) || undefined,
                    axisLabel: {
                        interval: 0,
                        overflow: 'break',
                        hideOverlap: true,
                        rotate: yAxis.rotateLabel,
                        formatter: (value) => formatChartValue(value, yAxis.formatting),
                    },
                    axisLine: {
                        show: false,
                        onZero: false,
                    },
                    splitLine: {
                        lineStyle: {
                            color: [themeVariables['extra-light']],
                        },
                    },
                    nameTextStyle: {
                        align: labelPosition === 'center' ? 'center' : position,
                    },
                };
                // If we provide undefined, log scale breaks
                if (tempYAxis.type === 'log') {
                    delete tempYAxis.min;
                    delete tempYAxis.max;
                }
                if (horizontal) {
                    options.xAxis = [tempYAxis];
                    options.yAxis = [xAxis];
                }
                else {
                    options.xAxis = [xAxis];
                    options.yAxis = [tempYAxis];
                }
            }
        }
        options.series = datasets.map(({ formatting, type, label, data, stack, tooltip, fill, smooth, step, roseType, symbol }, index) => {
            const { fixed, relative } = tooltip;
            // We should render the first metric in the dataset as the last
            const z = (datasets.length - 1) - index;
            if (['pie', 'doughnut'].includes(type)) {
                const startRadius = type === 'doughnut' ? 40 : 0;
                const endRadius = 80;
                const radiusLength = (endRadius - startRadius) / (datasets.length || 1);
                const sr = startRadius + (index * radiusLength);
                const er = startRadius + ((index + 1) * radiusLength);
                options.tooltip.trigger = 'item';
                let lbl = {
                    rotate: dimension.rotateLabel ? +dimension.rotateLabel : 0,
                };
                if (t === null || t === void 0 ? void 0 : t.labelsNextToPartition) {
                    lbl = Object.assign(Object.assign({}, lbl), { show: true, overflow: 'truncate' });
                }
                else {
                    lbl = Object.assign(Object.assign({}, lbl), { show: fixed, position: 'inside', align: 'center', verticalAlign: 'middle' });
                }
                return {
                    z,
                    stack,
                    name: label,
                    type: 'pie',
                    roseType,
                    radius: [`${sr}%`, `${er}%`],
                    center: ['50%', '55%'],
                    tooltip: {
                        trigger: 'item',
                        appendToBody: true,
                        formatter: (params) => {
                            const v = formatChartValue(params.value || '', formatting);
                            if (t === null || t === void 0 ? void 0 : t.formatting) {
                                return formatChartTooltip(t === null || t === void 0 ? void 0 : t.formatting, params);
                            }
                            return `${params.seriesName}<br>${params.marker}${params.name}<span style="float: right; margin-left: 20px">${v}${relative ? ' (' + params.percent + '%)' : ''}</span>`;
                        },
                    },
                    label: Object.assign(Object.assign({}, lbl), { formatter: (params) => formatChartValue(params.value || '', formatting) }),
                    itemStyle: {
                        borderRadius: 5,
                        borderColor: themeVariables.white,
                        borderWidth: 1,
                    },
                    emphasis: {
                        itemStyle: {
                            shadowBlur: 10,
                            shadowOffsetX: 0,
                            shadowColor: 'rgba(0, 0, 0, 0.5)',
                        },
                    },
                    data: labels.map((name, i) => {
                        return { name, value: data[i] };
                    }),
                    top: (offset === null || offset === void 0 ? void 0 : offset.isDefault) ? undefined : offset === null || offset === void 0 ? void 0 : offset.top,
                    right: (offset === null || offset === void 0 ? void 0 : offset.isDefault) ? undefined : offset === null || offset === void 0 ? void 0 : offset.right,
                    bottom: (offset === null || offset === void 0 ? void 0 : offset.isDefault) ? undefined : offset === null || offset === void 0 ? void 0 : offset.bottom,
                    left: (offset === null || offset === void 0 ? void 0 : offset.isDefault) ? undefined : offset === null || offset === void 0 ? void 0 : offset.left,
                };
            }
            else if (['bar', 'line', 'scatter'].includes(type)) {
                options.tooltip.trigger = 'axis';
                const defaultOffset = {
                    top: 65,
                    right: timeline.includes('x') ? 40 : 30,
                    bottom: timeline.includes('x') ? 60 : 20,
                    left: 30,
                };
                options.grid = {
                    top: (offset === null || offset === void 0 ? void 0 : offset.isDefault) ? defaultOffset.top : offset === null || offset === void 0 ? void 0 : offset.top,
                    right: (offset === null || offset === void 0 ? void 0 : offset.isDefault) ? defaultOffset.right : offset === null || offset === void 0 ? void 0 : offset.right,
                    bottom: (offset === null || offset === void 0 ? void 0 : offset.isDefault) ? defaultOffset.bottom : offset === null || offset === void 0 ? void 0 : offset.bottom,
                    left: (offset === null || offset === void 0 ? void 0 : offset.isDefault) ? defaultOffset.left : offset === null || offset === void 0 ? void 0 : offset.left,
                    containLabel: true,
                };
                if (horizontal) {
                    data = labels.map((name, i) => {
                        return [data[i], name];
                    });
                }
                else {
                    data = labels.map((name, i) => {
                        return [name, data[i]];
                    });
                }
                return {
                    z,
                    stack,
                    name: label,
                    type: type,
                    smooth,
                    step,
                    areaStyle: {
                        opacity: fill ? 0.7 : 0,
                    },
                    symbol,
                    symbolSize: type === 'scatter' ? 16 : 10,
                    tooltip: {
                        appendToBody: true,
                        // pass trigger type to determine if valueFormatter or formatter will be used
                        trigger: (t === null || t === void 0 ? void 0 : t.formatting) ? 'item' : 'axis',
                        // we can either
                        // add formatting to the value and apply tooltip if trigger: 'item'
                        // display the same tooltip format name <br/> seriesName value if trigger: 'axis'
                        // works when trigger is set to axis
                        valueFormatter: (value) => formatChartValue(value, formatting),
                        // works when trigger is set to item
                        formatter: (params) => {
                            const { value = [], percent = '' } = params;
                            const formattedValue = formatChartValue(value[1], formatting);
                            if (t === null || t === void 0 ? void 0 : t.formatting) {
                                return formatChartTooltip(t === null || t === void 0 ? void 0 : t.formatting, Object.assign(Object.assign({}, params), { value: value[1], percent }));
                            }
                            return `${params.seriesName}<br>${params.marker}${params.name}<span style="float: right; margin-left: 20px">${formattedValue}${relative ? ' (' + params.percent + '%)' : ''}</span>`;
                        },
                    },
                    label: {
                        show: fixed,
                        position: 'inside',
                        align: 'center',
                        verticalAlign: 'middle',
                        tooltip: {
                            trigger: 'axis',
                        },
                        formatter: (params) => {
                            const { value = [], percent = '' } = params;
                            return `${formatChartValue(value[1], formatting)}${relative ? ` (${percent}%)` : ''}`;
                        },
                    },
                    data,
                };
            }
        });
        const dataZoom = timeline ? [
            {
                show: timeline.includes('x'),
                type: 'slider',
                height: 30,
            },
            {
                show: timeline.includes('y'),
                type: 'slider',
                width: 15,
                yAxisIndex: 0,
            },
        ] : undefined;
        return Object.assign({ color: getColorschemeColors(colorScheme, data.customColorSchemes), textStyle: {
                fontFamily: themeVariables['font-regular'],
                overflow: 'break',
                color: themeVariables.black,
            }, toolbox: {
                feature: {
                    saveAsImage: saveAsImage ? {
                        name: this.name,
                    } : undefined,
                },
                top: 23,
                right: 2,
            }, dataZoom, legend: {
                show: !(l === null || l === void 0 ? void 0 : l.isHidden),
                type: (l === null || l === void 0 ? void 0 : l.isScrollable) ? 'scroll' : 'plain',
                top: (((_a = l === null || l === void 0 ? void 0 : l.position) === null || _a === void 0 ? void 0 : _a.isDefault) ? undefined : (_b = l === null || l === void 0 ? void 0 : l.position) === null || _b === void 0 ? void 0 : _b.top) || undefined,
                right: (((_c = l === null || l === void 0 ? void 0 : l.position) === null || _c === void 0 ? void 0 : _c.isDefault) ? undefined : (_d = l === null || l === void 0 ? void 0 : l.position) === null || _d === void 0 ? void 0 : _d.right) || undefined,
                bottom: (((_e = l === null || l === void 0 ? void 0 : l.position) === null || _e === void 0 ? void 0 : _e.isDefault) ? undefined : (_f = l === null || l === void 0 ? void 0 : l.position) === null || _f === void 0 ? void 0 : _f.bottom) || undefined,
                left: (((_g = l === null || l === void 0 ? void 0 : l.position) === null || _g === void 0 ? void 0 : _g.isDefault) ? (l === null || l === void 0 ? void 0 : l.align) || 'center' : (_h = l === null || l === void 0 ? void 0 : l.position) === null || _h === void 0 ? void 0 : _h.left) || 'auto',
                orient: (l === null || l === void 0 ? void 0 : l.orientation) || 'horizontal',
                textStyle: {
                    color: themeVariables.black,
                },
                pageTextStyle: {
                    color: themeVariables.black,
                },
                pageIconColor: themeVariables.black,
                pageIconInactiveColor: themeVariables.light,
            } }, options);
    }
    defMetric() {
        return Object.assign(super.defMetric(), {
            smooth: true,
            fill: false,
            rose: false,
            symbol: 'circle',
        });
    }
    baseChartType(datasets) {
        return datasets[0].type;
    }
}

class FunnelChart extends BaseChart {
    constructor(def = {}) {
        super(def);
        // Assure required fields; this helps with backwards compatibility
        for (const v of (this.config.reports || [])) {
            for (const d of (v.dimensions || [])) {
                if (!d.meta) {
                    d.meta = {};
                }
                if (!d.meta.fields) {
                    d.meta.fields = [];
                }
            }
            for (const m of (v.metrics || [])) {
                if (m.cumulative === undefined) {
                    m.cumulative = true;
                }
            }
        }
    }
    /**
     * Since funnel charts always define one type, this check can be simplified
     */
    mtrCheck({ field, aggregate }) {
        if (!field) {
            throw new Error('notification.chart.invalidConfig.missingMetricsField');
        }
        if (field !== 'count' && !aggregate) {
            throw new Error('notification.chart.invalidConfig.missingMetricsAggregate');
        }
    }
    /**
     * Extend this method to include filtering for just specific values.
     * For example:
     * We wish to show only new and converted leads.
     */
    formatReporterParams(r) {
        const base = super.formatReporterParams(r);
        const ff = base.filter;
        let df = '';
        if (r.dimensions && r.dimensions[0]) {
            const rd = r.dimensions[0];
            if (r.dimensions[0].meta) {
                const fields = r.dimensions[0].meta.fields || [];
                df = fields.map(({ value }) => `${rd.field || ''}='${value}'`)
                    .join(' OR ');
            }
        }
        if (ff && df) {
            base.filter = `(${base.filter}) AND (${df})`;
        }
        else if (!ff && df) {
            base.filter = df;
        }
        return base;
    }
    // Funnel chart creates a metric including all reports, so this step is deferred to there
    makeDataset(m, d, data, alias) {
        return {
            type: m.type,
            label: m.label || m.field,
            data,
            tooltip: {
                fixed: !!m.fixTooltips,
                relative: !!m.relativeValue,
            },
            formatting: m.formatting,
        };
    }
    makeOptions(data) {
        var _a, _b, _c, _d, _e, _f, _g, _h;
        const { reports = [], colorScheme, noAnimation = false, toolbox } = this.config;
        const { saveAsImage } = toolbox || {};
        const { labels, datasets = [], tooltip, themeVariables = {} } = data;
        const { fixed, relative } = tooltip;
        const { legend: l } = reports[0] || {};
        const { formatting } = datasets[0] || {};
        const colors = getColorschemeColors(colorScheme, data.customColorSchemes);
        return {
            animation: !noAnimation,
            textStyle: {
                fontFamily: themeVariables['font-regular'],
                overflow: 'break',
                color: themeVariables.black,
            },
            toolbox: {
                feature: {
                    saveAsImage: saveAsImage ? {
                        name: this.name,
                    } : undefined,
                },
                top: 23,
                right: 2,
            },
            tooltip: {
                trigger: 'item',
                formatter: (params) => {
                    const { value = '', percent = '' } = params;
                    const v = formatChartValue(value, formatting);
                    return `${params.seriesName}<br>${params.marker}${params.name}<span style="float: right; margin-left: 20px">${v}${relative ? ' (' + percent + '%)' : ''}</span>`;
                },
                appendToBody: true,
            },
            legend: {
                show: !(l === null || l === void 0 ? void 0 : l.isHidden),
                type: (l === null || l === void 0 ? void 0 : l.isScrollable) ? 'scroll' : 'plain',
                top: (((_a = l === null || l === void 0 ? void 0 : l.position) === null || _a === void 0 ? void 0 : _a.isDefault) ? undefined : (_b = l === null || l === void 0 ? void 0 : l.position) === null || _b === void 0 ? void 0 : _b.top) || undefined,
                right: (((_c = l === null || l === void 0 ? void 0 : l.position) === null || _c === void 0 ? void 0 : _c.isDefault) ? undefined : (_d = l === null || l === void 0 ? void 0 : l.position) === null || _d === void 0 ? void 0 : _d.right) || undefined,
                bottom: (((_e = l === null || l === void 0 ? void 0 : l.position) === null || _e === void 0 ? void 0 : _e.isDefault) ? undefined : (_f = l === null || l === void 0 ? void 0 : l.position) === null || _f === void 0 ? void 0 : _f.bottom) || undefined,
                left: (((_g = l === null || l === void 0 ? void 0 : l.position) === null || _g === void 0 ? void 0 : _g.isDefault) ? (l === null || l === void 0 ? void 0 : l.align) || 'center' : (_h = l === null || l === void 0 ? void 0 : l.position) === null || _h === void 0 ? void 0 : _h.left) || 'auto',
                orient: (l === null || l === void 0 ? void 0 : l.orientation) || 'horizontal',
                textStyle: {
                    color: themeVariables.black,
                },
                pageTextStyle: {
                    color: themeVariables.black,
                },
                pageIconColor: themeVariables.black,
                pageIconInactiveColor: themeVariables.light,
            },
            series: datasets.map(({ data, label, formatting }) => {
                return {
                    name: label,
                    type: 'funnel',
                    sort: 'descending',
                    top: 45,
                    bottom: 10,
                    left: '5%',
                    width: '90%',
                    label: {
                        show: fixed,
                        position: 'inside',
                        align: 'center',
                        verticalAlign: 'middle',
                        formatter: (params) => {
                            const { value = '', percent = '' } = params;
                            const formattedValue = formatChartValue(value, formatting);
                            return `${formattedValue}${relative ? ' (' + percent + '%)' : ''}`;
                        },
                    },
                    emphasis: {
                        label: {
                            show: fixed,
                            fontSize: 14,
                        },
                    },
                    data: labels.map((name, i) => {
                        return { name, value: data[i], itemStyle: { color: colors[i] } };
                    }),
                };
            }),
        };
    }
    baseChartType() {
        return 'funnel';
    }
    /**
     * Includes a few additional post processing steps:
     * * generate a set of labels based on all reports, all data sets,
     * * generates a set of data based on all reports, all data sets,
     */
    fetchReports(a) {
        const _super = Object.create(null, {
            fetchReports: { get: () => super.fetchReports }
        });
        return __awaiter(this, void 0, void 0, function* () {
            var _a, _b, _c;
            const rr = yield _super.fetchReports.call(this, a);
            const values = [];
            let tooltip = {};
            let label = '';
            let formatting = {};
            // Above provided data sets might not have their labels/values ordered
            // correctly
            const valMap = {};
            // Map values to their labels
            for (let ri = 0; ri < rr.length; ri++) {
                const r = rr[ri];
                r.labels.forEach((l, i) => {
                    valMap[l] = r.datasets[0].data[i];
                });
                tooltip = Object.assign(Object.assign({}, tooltip), r.datasets[0].tooltip);
                label = r.datasets[0].label;
                formatting = r.datasets[0].formatting;
                // Construct labels & data based on provided reports
                const report = (_a = this.config.reports) === null || _a === void 0 ? void 0 : _a[ri];
                const d = (_b = report === null || report === void 0 ? void 0 : report.dimensions) === null || _b === void 0 ? void 0 : _b[0];
                let { fields = [] } = d.meta || {};
                fields = fields.length ? fields : r.labels;
                for (const label of fields) {
                    const value = typeof label === 'object' ? label.value : label;
                    values.push({
                        // Use value for label and resolve it on FE (i18n)
                        label: value,
                        data: valMap[value] || 0,
                    });
                }
            }
            // We are rendering the chart upside down
            // (by default it renders in ASC, but we want DESC)
            const labels = [];
            const data = [];
            values.sort((a, b) => a.data - b.data).forEach(v => {
                labels.push(v.label);
                data.push(v.data);
            });
            (_c = this.config.reports) === null || _c === void 0 ? void 0 : _c.forEach(r => {
                var _a, _b;
                const dimension = (_a = r.dimensions) === null || _a === void 0 ? void 0 : _a[0];
                if ((_b = dimension === null || dimension === void 0 ? void 0 : dimension.meta) === null || _b === void 0 ? void 0 : _b.fields) {
                    for (const { value, color } of dimension.meta.fields) {
                    }
                }
            });
            // Get cumulative data but also keep original for tooltips
            if (this.isCumulative()) {
                for (let i = 1; i < data.length; i++) {
                    data[i] += data[i - 1];
                }
            }
            return {
                labels,
                datasets: [{
                        label,
                        data,
                        formatting,
                    }],
                tooltip,
            };
        });
    }
    isCumulative() {
        // Cumulative true by default
        // Find false value
        let cumulative = true;
        const { reports = [] } = this.config;
        reports.forEach(({ metrics = [] }) => {
            if (cumulative && !metrics[0].cumulative) {
                cumulative = false;
            }
        });
        return cumulative;
    }
    defMetric() {
        return Object.assign(super.defMetric(), {
            type: ChartType.funnel,
            fixTooltips: false,
            relativeValue: true,
        });
    }
    defDimension() {
        return Object.assign({}, {
            conditions: {},
            meta: { fields: [] },
        });
    }
}

class GaugeChart extends BaseChart {
    constructor(def = {}) {
        super(def);
        // Assure required fields
        for (const v of (this.config.reports || [])) {
            for (const d of (v.dimensions || [])) {
                // Since gauge produces one value we want one dataset, deletedAt is the same for all existing records
                d.field = 'deletedAt';
                if (!d.meta) {
                    d.meta = {};
                }
                if (!d.meta.steps) {
                    d.meta.steps = [];
                }
            }
        }
    }
    // Gauge charts (at the moment) support only 1 report per chart
    fetchReports(a) {
        const _super = Object.create(null, {
            fetchReports: { get: () => super.fetchReports }
        });
        return __awaiter(this, void 0, void 0, function* () {
            return _super.fetchReports.call(this, a).then((rr) => {
                return rr[0];
            });
        });
    }
    processLabels(ll, d) {
        var _a;
        return (((_a = d.meta) === null || _a === void 0 ? void 0 : _a.steps) || []).map(({ label }) => label);
    }
    makeDataset(m, d, data, alias) {
        var _a;
        const steps = (((_a = d.meta) === null || _a === void 0 ? void 0 : _a.steps) || []);
        data = this.datasetPostProc(data, m);
        const value = data.reduce((acc, cur) => {
            return !isNaN(cur) ? acc + parseFloat(cur) : acc;
        }, 0).toFixed(3);
        const max = Math.max(...steps.map(({ value }) => parseFloat(value)));
        const sortedSteps = [...steps].sort((a, b) => {
            return parseFloat(b.value) - parseFloat(a.value);
        });
        const { label: name } = sortedSteps.reduce((acc, cur) => {
            const curValue = parseFloat(cur.value);
            return value < curValue ? cur : acc;
        }, sortedSteps[0] || {});
        return {
            steps,
            name,
            max,
            value,
            startAngle: m.startAngle,
            endAngle: m.endAngle,
            tooltip: {
                fixed: m.fixTooltips,
            },
            formatting: m.formatting,
        };
    }
    makeOptions(data) {
        const { reports = [], colorScheme, noAnimation = false, toolbox } = this.config;
        const { saveAsImage } = toolbox || {};
        const { datasets = [], themeVariables = {} } = data;
        const { steps = [], name, value, max, tooltip, startAngle, endAngle, formatting, } = datasets.find(({ value }) => value) || datasets[0];
        const colors = getColorschemeColors(colorScheme, data.customColorSchemes);
        const color = steps.map((s, i) => {
            return [s.value / max, colors[i]];
        });
        return {
            animation: !noAnimation,
            textStyle: {
                fontFamily: themeVariables['font-regular'],
                overflow: 'break',
                color: themeVariables.black,
            },
            toolbox: {
                feature: {
                    saveAsImage: saveAsImage ? {
                        name: this.name,
                    } : undefined,
                },
                top: 23,
                right: 2,
            },
            grid: {
                bottom: 0,
            },
            series: [
                {
                    type: 'gauge',
                    startAngle,
                    endAngle,
                    min: 0,
                    max,
                    splitNumber: 5,
                    radius: '100%',
                    center: ['50%', '50%'],
                    pointer: {
                        width: 5,
                        length: '75%',
                        itemStyle: {
                            color: themeVariables.black,
                        },
                    },
                    splitLine: {
                        distance: 0,
                        length: 0,
                        lineStyle: {
                            color: themeVariables.white,
                        },
                    },
                    axisLine: {
                        lineStyle: {
                            width: 30,
                            color,
                        },
                    },
                    axisTick: {
                        show: false,
                        distance: -30,
                    },
                    axisLabel: {
                        show: false,
                        distance: 60,
                    },
                    title: {
                        fontSize: 14,
                        show: tooltip.fixed,
                        offsetCenter: [0, '30%'],
                        color: themeVariables.black,
                    },
                    detail: {
                        fontSize: 13,
                        offsetCenter: [0, '55%'],
                        valueAnimation: true,
                        color: themeVariables.black,
                        formatter: (value) => formatChartValue(value, formatting),
                    },
                    data: [
                        {
                            name,
                            value,
                        },
                    ],
                },
            ],
        };
    }
    baseChartType() {
        return 'gauge';
    }
    defMetric() {
        return Object.assign(super.defMetric(), {
            type: ChartType.gauge,
            fixTooltips: true,
            startAngle: 200,
            endAngle: -20,
        });
    }
    /**
     * Checks validity of dimensions.
     * If invalid it throws an error
     */
    dimCheck({ meta }) {
        if (((meta === null || meta === void 0 ? void 0 : meta.steps) || []).length === 0) {
            throw new Error('notification.chart.invalidConfig.missingDimensionsSteps');
        }
    }
    /**
     * Since gauge charts always define one type, this check can be simplified
     */
    mtrCheck({ field, aggregate }) {
        if (!field) {
            throw new Error('notification.chart.invalidConfig.missingMetricsField');
        }
        if (field !== 'count' && !aggregate) {
            throw new Error('notification.chart.invalidConfig.missingMetricsAggregate');
        }
    }
}

class RadarChart extends BaseChart {
    mtrCheck({ field, aggregate }) {
        if (!field) {
            throw new Error('notification.chart.invalidConfig.missingMetricsField');
        }
        if (field !== 'count' && !aggregate) {
            throw new Error('notification.chart.invalidConfig.missingMetricsAggregate');
        }
    }
    makeDataset(m, d, data) {
        return {
            type: m.type,
            label: m.label || m.field,
            data,
            formatting: m.formatting,
        };
    }
    makeOptions(data) {
        var _a, _b, _c, _d, _e, _f, _g, _h;
        const { reports = [], colorScheme, noAnimation = false, toolbox } = this.config;
        const { saveAsImage } = toolbox || {};
        const { labels, datasets = [], dimension = {}, themeVariables = {} } = data;
        const { legend: l, } = reports[0] || {};
        const { formatting } = datasets[0] || {};
        let min = 0;
        let max = Math.max();
        const seriesData = [];
        datasets.forEach(({ data: value, label: name }) => {
            value.forEach((v) => {
                if (v < min)
                    min = v;
                if (v > max)
                    max = v;
            });
            seriesData.push({ value, name });
        });
        return {
            color: getColorschemeColors(colorScheme, data.customColorSchemes),
            animation: !noAnimation,
            textStyle: {
                fontFamily: themeVariables['font-regular'],
                overflow: 'break',
                color: themeVariables.black,
            },
            toolbox: {
                feature: {
                    saveAsImage: saveAsImage ? {
                        name: this.name,
                    } : undefined,
                },
                top: 23,
                right: 2,
            },
            legend: {
                show: !(l === null || l === void 0 ? void 0 : l.isHidden),
                type: (l === null || l === void 0 ? void 0 : l.isScrollable) ? 'scroll' : 'plain',
                top: (((_a = l === null || l === void 0 ? void 0 : l.position) === null || _a === void 0 ? void 0 : _a.isDefault) ? undefined : (_b = l === null || l === void 0 ? void 0 : l.position) === null || _b === void 0 ? void 0 : _b.top) || undefined,
                right: (((_c = l === null || l === void 0 ? void 0 : l.position) === null || _c === void 0 ? void 0 : _c.isDefault) ? undefined : (_d = l === null || l === void 0 ? void 0 : l.position) === null || _d === void 0 ? void 0 : _d.right) || undefined,
                bottom: (((_e = l === null || l === void 0 ? void 0 : l.position) === null || _e === void 0 ? void 0 : _e.isDefault) ? undefined : (_f = l === null || l === void 0 ? void 0 : l.position) === null || _f === void 0 ? void 0 : _f.bottom) || undefined,
                left: (((_g = l === null || l === void 0 ? void 0 : l.position) === null || _g === void 0 ? void 0 : _g.isDefault) ? (l === null || l === void 0 ? void 0 : l.align) || 'center' : (_h = l === null || l === void 0 ? void 0 : l.position) === null || _h === void 0 ? void 0 : _h.left) || 'auto',
                orient: (l === null || l === void 0 ? void 0 : l.orientation) || 'horizontal',
                textStyle: {
                    color: themeVariables.black,
                },
                pageTextStyle: {
                    color: themeVariables.black,
                },
                pageIconColor: themeVariables.black,
                pageIconInactiveColor: themeVariables.light,
            },
            tooltip: {
                show: true,
                position: 'top',
                appendToBody: true,
                valueFormatter: (value) => formatChartValue(value, formatting),
            },
            radar: {
                shape: dimension.shape,
                indicator: labels.map((name) => {
                    return { name, min, max };
                }),
                center: ['50%', '55%'],
            },
            series: {
                type: 'radar',
                label: {
                    show: dimension.fixTooltips,
                    formatter: (params) => {
                        const { value = '' } = params;
                        return formatChartValue(value, formatting);
                    },
                },
                data: seriesData,
            },
        };
    }
    baseChartType() {
        return 'radar';
    }
    fetchReports(a) {
        const _super = Object.create(null, {
            fetchReports: { get: () => super.fetchReports }
        });
        return __awaiter(this, void 0, void 0, function* () {
            return _super.fetchReports.call(this, a).then((rr) => {
                return rr[0];
            });
        });
    }
    defMetric() {
        return Object.assign(super.defMetric(), {
            type: ChartType.radar,
        });
    }
    defDimension() {
        return Object.assign({}, {
            shape: 'polygon',
            fixTooltips: false,
            conditions: {},
            meta: {},
        });
    }
}

function namespaceMatcher(r, c, def) {
    if (!r) {
        throw new Error('can not run namespace matcher on undefined/null namespace');
    }
    // keep in sync with server/compose/service/event/namespace.go
    switch (c.Name()) {
        case 'namespace':
        case 'namespace.slug':
            return c.Match(r.slug);
        case 'namespace.name':
            return c.Match(r.name);
    }
    return def;
}
function moduleMatcher(r, c, def) {
    if (!r) {
        throw new Error('can not run module matcher on undefined/null module');
    }
    // keep in sync with server/compose/service/event/module.go
    switch (c.Name()) {
        case 'module':
        case 'module.handle':
            return c.Match(r.handle);
        case 'module.name':
            return c.Match(r.name);
    }
    return def;
}
/**
 * Creates event for compose resource with ready-to-go-defaults
 */
function ComposeEvent(event) {
    return Object.assign({ eventType: onManual, resourceType: 'compose', match: () => true }, event);
}
/**
 * Creates namespace event with ready-to-go-defaults
 */
function NamespaceEvent(res, event) {
    return Object.assign(Object.assign({ eventType: onManual, resourceType: res.resourceType, match: (c) => namespaceMatcher(res, c, false) }, event), { 
        // Override the arguments at the end
        args: Object.assign({ namespace: res }, event === null || event === void 0 ? void 0 : event.args) });
}
/**
 * Creates module event with ready-to-go-defaults
 */
function ModuleEvent(res, event) {
    return Object.assign(Object.assign({ eventType: onManual, resourceType: res.resourceType, match: (c) => namespaceMatcher(res.namespace, c, moduleMatcher(res, c, false)) }, event), { 
        // Override the arguments at the end
        args: Object.assign({ module: res, namespace: res.namespace }, event === null || event === void 0 ? void 0 : event.args) });
}
/**
 * Creates record event with ready-to-go-defaults
 */
function RecordEvent(res, event) {
    return Object.assign(Object.assign({ eventType: onManual, resourceType: res.resourceType, match: (c) => namespaceMatcher(res.namespace, c, moduleMatcher(res.module, c, false)) }, event), { 
        // Override the arguments at the end
        args: Object.assign({ record: res, module: res.module, namespace: res.namespace }, event === null || event === void 0 ? void 0 : event.args) });
}
/**
 * Creates record event with ready-to-go-defaults
 */
function PageEvent(res, event) {
    return Object.assign(Object.assign({ eventType: onManual, resourceType: 'compose:page', match: () => true }, event), { 
        // Override the arguments at the end
        args: Object.assign({ page: res }, event === null || event === void 0 ? void 0 : event.args) });
}
/**
 * Returns handler that routes onManual events for server script to the compose API
 *
 * See makeAutomationScriptsRegistrator
 *
 * @param api
 * @return function
 */
function TriggerComposeServerScriptOnManual(api) {
    return (ev, script) => {
        const params = { script, args: ev.args };
        if (ev.resourceType === 'compose') {
            return api
                .automationTriggerScript(Object.assign({}, params));
        }
        if (!ev.args) {
            throw new Error('expecting args prop in event');
        }
        if (ev.resourceType === 'compose:namespace') {
            if (!IsOf(ev.args.namespace, 'namespaceID')) {
                throw new Error('expecting args.namespace in event arguments');
            }
            const { namespaceID } = ev.args.namespace;
            return api
                .namespaceTriggerScript(Object.assign({ namespaceID }, params));
        }
        if (ev.resourceType === 'compose:module') {
            if (!IsOf(ev.args.module, 'namespaceID', 'moduleID')) {
                throw new Error('expecting args.module in event arguments');
            }
            const { namespaceID, moduleID } = ev.args.module;
            return api
                .moduleTriggerScript(Object.assign({ namespaceID, moduleID }, params));
        }
        if (ev.resourceType === 'compose:record') {
            if (!IsOf(ev.args.record, 'namespaceID', 'moduleID', 'recordID')) {
                throw new Error('expecting args.record in event arguments');
            }
            const record = ev.args.record;
            const { namespaceID, moduleID, recordID, values } = record;
            return api
                .recordTriggerScript(Object.assign({ namespaceID, moduleID, recordID, values }, params))
                .then(rval => record.apply(rval));
        }
        throw Error(`cannot trigger server script: unknown resource type '${ev.resourceType}'`);
    };
}

var index$3 = /*#__PURE__*/Object.freeze({
  __proto__: null,
  Chart: Chart,
  ComposeEvent: ComposeEvent,
  FunnelChart: FunnelChart,
  GaugeChart: GaugeChart,
  Module: Module,
  ModuleEvent: ModuleEvent,
  ModuleField: ModuleField,
  ModuleFieldBool: ModuleFieldBool,
  ModuleFieldDateTime: ModuleFieldDateTime,
  ModuleFieldEmail: ModuleFieldEmail,
  ModuleFieldFile: ModuleFieldFile,
  ModuleFieldGeometry: ModuleFieldGeometry,
  ModuleFieldMaker: ModuleFieldMaker,
  ModuleFieldNumber: ModuleFieldNumber,
  ModuleFieldRecord: ModuleFieldRecord,
  ModuleFieldRegistry: Registry$1,
  ModuleFieldSelect: ModuleFieldSelect,
  ModuleFieldString: ModuleFieldString,
  ModuleFieldUrl: ModuleFieldUrl,
  ModuleFieldUser: ModuleFieldUser,
  Namespace: Namespace,
  NamespaceEvent: NamespaceEvent,
  Page: Page,
  PageBlock: PageBlock,
  PageBlockAutomation: PageBlockAutomation,
  PageBlockCalendar: PageBlockCalendar,
  PageBlockChart: PageBlockChart,
  PageBlockComment: PageBlockComment,
  PageBlockContent: PageBlockContent,
  PageBlockFile: PageBlockFile,
  PageBlockGeometry: PageBlockGeometry,
  PageBlockIFrame: PageBlockIFrame,
  PageBlockMaker: PageBlockMaker,
  PageBlockMetric: PageBlockMetric,
  PageBlockNavigation: PageBlockNavigation,
  PageBlockProgress: PageBlockProgress,
  PageBlockRecord: PageBlockRecord,
  PageBlockRecordList: PageBlockRecordList,
  PageBlockRecordOrganizer: PageBlockRecordOrganizer,
  PageBlockRecordRevisions: PageBlockRecordRevisions,
  PageBlockRegistry: Registry,
  PageBlockReport: PageBlockReport,
  PageBlockSocialFeed: PageBlockSocialFeed,
  PageBlockTab: PageBlockTab,
  PageEvent: PageEvent,
  PageLayout: PageLayout,
  RadarChart: RadarChart,
  Record: Record,
  RecordEvent: RecordEvent,
  RecordValidator: RecordValidator,
  TriggerComposeServerScriptOnManual: TriggerComposeServerScriptOnManual,
  chartUtil: chartUtil,
  convertRevisionPayloadToRevision: convertRevisionPayloadToRevision,
  getModuleFromYaml: getModuleFromYaml
});

const emailStyle = `
body { -ms-text-size-adjust: 100%; -webkit-text-size-adjust: 100%; color: #3A393C; font-family: Verdana,Arial,sans-serif; font-size: 14px; height: 100%; margin: 0; padding: 0; width: 100% !important; }
table { margin: 20px auto; background: #FFFFFF; border-collapse: collapse; max-width: 100%; }
table tr { height: 40px; }
table td { padding-top: 10px; padding-left: 20px; width:100%; max-width:100%; min-width:100%; width:100%; vertical-align: top; }
table tbody { border-top: 3px solid #808080; }
tbody tr:nth-child(even) { background-color: #F3F3F4; }
tbody td:first-child { width: 30%; color: #808080; }
tbody td:nth-child(2) { width: 70%; }
h2, p { padding: 10px 20px; }
p { text-align: justify; line-height: 1.4;}
`;
/**
 * Helpers to determine if specific object looks like the type we are interested in.
 * It does not rely on instanceof, because of bundling issues.
 */
function isRecord(o) {
    return o && !!o.recordID && o.moduleID && o.namespaceID;
}
function isModule(o) {
    return o && !!o.moduleID && o.namespaceID;
}
function isPage(o) {
    return o && !!o.pageID && o.namespaceID;
}
/**
 * ComposeHelper provides layer over Compose API and utilities that simplify automation script writing
 *
 * Initiated as Compose object and provides a few handy shortcuts and fallback that will enable you
 * to rapidly develop your automation scripts.
 */
class ComposeHelper {
    /**
     * @param ctx.$namespace - Current namespace
     * @param ctx.$module - Current module
     * @param ctx.$record - Current record
     */
    constructor(ctx) {
        this.ComposeAPI = ctx.ComposeAPI;
        this.$namespace = ctx.$namespace;
        this.$module = ctx.$module;
        this.$record = ctx.$record;
    }
    /**
     * Creates new Page object
     *
     * <p>
     *   Created page is "in-memory" only. To store it, use savePage() method
     * </p>
     *
     * @example
     * // Simple page creation new page on current namespace
     * let myPage = await Compose.makePage({ title: 'My Amazing Page!' })
     *
     * @param values
     * @param ns - defaults to current $namespace
     */
    makePage() {
        return __awaiter(this, arguments, void 0, function* (values = {}, ns = this.$namespace) {
            return this.resolveNamespace(ns).then(ns => {
                return new Page(Object.assign(Object.assign({}, values), { namespaceID: ns.namespaceID }));
            });
        });
    }
    /**
     * Creates/updates Page
     *
     * @param page
     */
    savePage(page) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(page).then(page => {
                if (!isPage(page)) {
                    throw Error('expecting Page type');
                }
                if (page.pageID && isFresh(page.pageID)) {
                    return this.ComposeAPI.pageCreate(kv(page)).then(page => new Page(page));
                }
                else {
                    return this.ComposeAPI.pageUpdate(kv(page)).then(page => new Page(page));
                }
            });
        });
    }
    /**
     * Deletes a page
     *
     * @example
     * Compose.deletePage(myPage)
     *
     * @param page
     */
    deletePage(page) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(page).then(page => {
                if (!isPage(page)) {
                    throw Error('expecting Page type');
                }
                if (!isFresh(page.pageID)) {
                    return this.ComposeAPI.pageDelete(kv(page));
                }
            });
        });
    }
    /**
     * Searches for pages
     *
     * @private
     * @param filter
     * @param ns
     */
    findPages() {
        return __awaiter(this, arguments, void 0, function* (filter = {}, ns = this.$namespace) {
            if (typeof filter === 'string') {
                filter = { query: filter };
            }
            return this.resolveNamespace(ns).then(ns => {
                const namespaceID = extractID(ns, 'namespaceID');
                return this.ComposeAPI.pageList(Object.assign({ namespaceID }, filter)).then(res => {
                    // Casting all we got to to Page
                    res.set = res.set.map(m => new Page(m));
                    return res;
                });
            });
        });
    }
    /**
     * Finds page by ID
     *
     * @example
     * // Explicitly load page and do something with it
     * Compose.finePageByID('2039248239042').then(myPage => {
     *   // do something with myPage
     *   myPage.title = 'My More Amazing Page!'
     *   return myPage
     * }).then(Compose.savePage)
     *
     * @param page - accepts Page, pageID (when string string)
     * @param ns - namespace, defaults to current $namespace
     */
    findPageByID(page_1) {
        return __awaiter(this, arguments, void 0, function* (page, ns = this.$namespace) {
            return this.resolveNamespace(ns).then((ns) => {
                const pageID = extractID(page, 'pageID');
                const namespaceID = extractID(ns, 'namespaceID');
                return this.ComposeAPI.pageRead({ namespaceID, pageID }).then(page => new Page(page));
            });
        });
    }
    /**
     * Creates new Record object
     *
     * <p>
     *   Created record is "in-memory" only. To store it, use saveRecord() method
     * </p>
     *
     * @example
     * // Simple record creation (new record of current module - $module)
     * let myLead = await Compose.makeRecord()
     * myLead.values.Title = 'My Lead Title'
     *
     * // Create record of type Lead and copy values from another Record
     * // This will copy only values that have the same name in both modules
     * let myLead = await Compose.makeRecord(myContact, 'Lead')
     *
     * // Or use promises:
     * Compose.makeRecord(myContact, 'Lead').then(myLead => {
     *   myLead.values.Title = 'My Lead Title'
     *
     *   // ...
     *
     *   // return record when finished
     *   return myLead
     * }).catch(err => {
     *   // solve the problem
     *   console.error(err)
     * })
     *
     * @param values
     * @param module - defaults to current $module
     */
    makeRecord() {
        return __awaiter(this, arguments, void 0, function* (values = {}, module = null) {
            return this.resolveModule(module, this.$module).then(module => {
                const record = new Record(module);
                // Set record values
                record.setValues(values);
                return record;
            });
        });
    }
    /**
     * Saves a record
     *
     * Please note that there is no need to explicitly save (current record) on before/after events,
     * internal systems take care of that.
     *
     * @example
     * // Quick example how to make and save new Lead:
     * let mySavedLead = await Compose.saveRecord(Compose.makeRecord({Title: 'Lead title'}, 'Lead'))
     * if (mySavedLead) {
     *   console.log('Record saved, new ID', mySavedLead.recordID)
     * } else {
     *   // solve the problem
     *   console.error(err)
     * }
     *
     * // Or with promises:
     * Compose.makeRecord({Title: 'Lead title'}, 'Lead')).then(myLead => {
     *   return Compose.saveRecord(myLead)
     * }).then(mySavedLead => {
     *   console.log('Record saved, new ID', mySavedLead.recordID)
     * }).catch(err => {
     *   // solve the problem
     *   console.error(err)
     * })
     *
     * @param record
     */
    saveRecord(record) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(record).then(record => {
                if (!isRecord(record)) {
                    throw Error('expecting Record type');
                }
                if (isFresh(record.recordID)) {
                    return this.ComposeAPI.recordCreate(kv(record)).then(r => new Record(record.module, r));
                }
                else {
                    return this.ComposeAPI.recordUpdate(kv(record)).then(r => new Record(record.module, r));
                }
            });
        });
    }
    /**
     * Deletes a record
     *
     * Please note that there is no need to explicitly delete (current record) on before/after events.
     *
     * @example
     * Compose.deleteRecord(myLead)
     *
     * @param record
     */
    deleteRecord(record) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(record).then(record => {
                if (!isRecord(record)) {
                    throw Error('expecting Record type');
                }
                if (!isFresh(record.recordID)) {
                    return this.ComposeAPI.recordDelete(kv(record));
                }
            });
        });
    }
    /**
     * Searches for records of a specific record
     *
     * @example
     * // Find all records (of the current module)
     * Compose.findRecords()
     *
     * // Find Projects where ROI is more than 15%
     * // (assuming we have Project module with netProfit and totalInvestment numeric fields)
     * Compose.findRecords('netProfit / totalInvestment > 0.15', 'Project')
     *
     * // Find Projects where ROI is more than 15%
     * // (assuming we have Project module with netProfit and totalInvestment numeric fields)
     * Compose.findRecords('netProfit / totalInvestment > 0.15', 'Project')
     *
     * // More complex query with sorting:
     * // Returns top 5 Projects with more than 15% ROI in the last year
     * Compose.findRecords({
     *   filter: '(netProfit / totalInvestment > 0.15) AND (YEAR(createdAt) = YEAR(NOW()) - 1)'
     *   sort: 'netProfit / totalInvestment DESC',
     *   limit: 5,
     * }, 'Project')
     *
     * // Accessing returned records
     * Compose.findRecords().then(({ set, filter }) => {
     *    // set: array of records
     *    // filter: object with filter specs
     *
     *    Use internal Array functions
     *    set.forEach(r => {
     *      // r, one of the records each iteration
     *    })
     *
     *    // Or standard for-loop
     *    for (let r of set) {
     *       // r...
     *    }
     * })
     *
     * @param filter - filter object (or filtering conditions when string)
     * @property {string} filter.query - filtering conditions
     * @property {string} filter.sort - sorting rules
     * @property {number} filter.limit - number of max returned records
     * @property {number} filter.pageCursor - hashed string that retrieves a specific page
     * @param [module] - if not set, defaults to $module
     */
    findRecords() {
        return __awaiter(this, arguments, void 0, function* (filter = '', module = this.$module) {
            return this.resolveModule(module).then(module => {
                const { moduleID, namespaceID } = module;
                let params = {
                    moduleID,
                    namespaceID,
                };
                if (typeof filter === 'string') {
                    params.query = filter;
                }
                else if (typeof filter === 'object') {
                    params = Object.assign(Object.assign({}, params), filter);
                }
                return this.ComposeAPI.recordList(params).then(res => {
                    // Casting all we got to to Record
                    res.set = res.set.map(record => new Record(module, record));
                    return res;
                });
            });
        });
    }
    /**
     * Finds last (created) record in the module
     *
     * @example
     * Compose.findLastRecord('Settings').then(lastSettingRecord => {
     *   // handle lastSettingRecord
     * })
     *
     * @param module
     */
    findLastRecord() {
        return __awaiter(this, arguments, void 0, function* (module = this.$module) {
            return this.findRecords({ sort: 'createdAt DESC', limit: 1 }, module).then(res => {
                if (!Array.isArray(res.set) || res.set.length === 0) {
                    throw new Error('records not found');
                }
                return res.set[0];
            });
        });
    }
    /**
     * Finds first (created) record in the module
     *
     * @example
     * Compose.findFirstRecord('Settings').then(firstSettingRecord => {
     *   // handle this firstSettingRecord
     * })
     *
     * @param module
     */
    findFirstRecord() {
        return __awaiter(this, arguments, void 0, function* (module = this.$module) {
            return this.findRecords({ sort: 'createdAt', limit: 1 }, module).then(res => {
                if (!Array.isArray(res.set) || res.set.length === 0) {
                    throw new Error('records not found');
                }
                return res.set[0];
            });
        });
    }
    /**
     * Finds one record by ID
     *
     * @example
     * Compose.findRecordByID("23957823957").then(specificRecord => {
     *   // handle this specificRecord
     * })
     *
     * @param record
     * @param module
     */
    findRecordByID(record_1) {
        return __awaiter(this, arguments, void 0, function* (record, module = null) {
            // We're handling module default a bit differently here
            // because we want to allow users to use record's module
            return this.resolveModule(module, (record || {}).module, this.$module).then((module) => {
                const { moduleID, namespaceID } = module;
                return this.ComposeAPI.recordRead({
                    moduleID,
                    namespaceID,
                    recordID: extractID(record, 'recordID'),
                }).then(r => new Record(module, r));
            });
        });
    }
    /**
     * Finds a single attachment
     *
     * @param attachment Attachment to find
     * @param ns
     */
    findAttachmentByID(attachment_1) {
        return __awaiter(this, arguments, void 0, function* (attachment, ns = this.$namespace) {
            return this.resolveNamespace(ns).then(namespace => {
                const { namespaceID } = namespace;
                return this.ComposeAPI.attachmentRead({
                    kind: 'original',
                    attachmentID: extractID(attachment, 'attachmentID'),
                    namespaceID,
                }).then(att => new Attachment(att));
            });
        });
    }
    /**
     * Helper to determine field's name from it's label
     * @param label Field's label
     */
    moduleFieldNameFromLabel(label) {
        return label.split(/[^a-zA-Z0-9_]/g)
            .filter(p => !!p)
            .map(p => `${p[0].toUpperCase()}${p.slice(1)}`)
            .join('');
    }
    /**
     * Creates new Module object
     *
     * @param module
     * @param ns, defaults to current $namespace
     */
    makeModule() {
        return __awaiter(this, arguments, void 0, function* (module = {}, ns = this.$namespace) {
            return this.resolveNamespace(ns).then((ns) => {
                return new Module(Object.assign(Object.assign({}, module), { namespaceID: ns.namespaceID }));
            });
        });
    }
    /**
     * Creates/updates Module
     *
     * @param module
     */
    saveModule(module) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(module).then(module => {
                if (!isModule(module)) {
                    throw new Error('expecting Module type');
                }
                if (isFresh(module.moduleID)) {
                    return this.ComposeAPI.moduleCreate(kv(module)).then(m => new Module(m));
                }
                else {
                    return this.ComposeAPI.moduleUpdate(kv(module)).then(m => new Module(m));
                }
            });
        });
    }
    /**
     * Searches for modules
     *
     * @private
     * @param filter
     * @param ns
     */
    findModules() {
        return __awaiter(this, arguments, void 0, function* (filter = '', ns = this.$namespace) {
            if (typeof filter === 'string') {
                filter = { query: filter };
            }
            return this.resolveNamespace(ns).then((ns) => {
                const namespaceID = extractID(ns, 'namespaceID');
                return this.ComposeAPI.moduleList(Object.assign({ namespaceID }, filter)).then(res => {
                    // Casting all we got to to Module
                    res.set = res.set.map(m => new Module(m));
                    return res;
                });
            });
        });
    }
    /**
     * Finds module by ID
     *
     * @example
     * // Explicitly load module and do something with it
     * Compose.findModuleByID('2039248239042').then(myModule => {
     *   // do something with myModule
     *   return Compose.findLastRecord(myModule)
     * }).then((lastRecord) => {})
     *
     * // or
     * Compose.findLastRecord(Compose.findModuleByID('2039248239042')).then(....)
     *
     * // even shorter
     * Compose.findLastRecord('2039248239042').then(....)
     *
     * @param module - accepts Module, moduleID (when string) or Record
     * @param ns - namespace, defaults to current $namespace
     */
    findModuleByID(module_1) {
        return __awaiter(this, arguments, void 0, function* (module, ns = this.$namespace) {
            return this.resolveNamespace(ns).then((ns) => {
                const moduleID = extractID(module, 'moduleID');
                const namespaceID = extractID(ns, 'namespaceID');
                return this.ComposeAPI.moduleRead({ namespaceID, moduleID }).then(m => new Module(m));
            });
        });
    }
    /**
     * Finds module by name
     *
     * @example
     * // Explicitly load module and do something with it
     * Compose.findModuleByName('SomeModule').then(myModule => {
     *   // do something with myModule
     *   return Compose.findLastRecord(myModule)
     * }).then((lastRecord) => {})
     *
     * // or
     * Compose.findLastRecord(Compose.findModuleByName('SomeModule')).then(....)
     *
     * // even shorter
     * Compose.findLastRecord('SomeModule').then(....)
     *
     * @param name - name of the module
     * @param ns - defaults to current $namespace
     */
    findModuleByName(name_1) {
        return __awaiter(this, arguments, void 0, function* (name, ns = this.$namespace) {
            return this.resolveNamespace(ns).then((ns) => {
                const namespaceID = extractID(ns, 'namespaceID');
                return this.ComposeAPI.moduleList({ namespaceID, name }).then(res => {
                    if (!Array.isArray(res.set) || res.set.length === 0) {
                        throw new Error('module not found');
                    }
                    return new Module(res.set[0]);
                });
            });
        });
    }
    /**
     * Finds module by handle
     *
     * @example
     * // Explicitly load module and do something with it
     * Compose.findModuleByHandle('SomeModule').then(myModule => {
     *   // do something with myModule
     *   return Compose.findLastRecord(myModule)
     * }).then((lastRecord) => {})
     *
     * // or
     * Compose.findLastRecord(Compose.findModuleByHandle('SomeModule')).then(....)
     *
     * // even shorter
     * Compose.findLastRecord('SomeModule').then(....)
     *
     * @param handle - handle of the module
     * @param ns - defaults to current $namespace
     */
    findModuleByHandle(handle_1) {
        return __awaiter(this, arguments, void 0, function* (handle, ns = this.$namespace) {
            return this.resolveNamespace(ns).then((ns) => {
                const namespaceID = extractID(ns, 'namespaceID');
                return this.ComposeAPI.moduleList({ namespaceID, handle }).then(res => {
                    if (!Array.isArray(res.set) || res.set.length === 0) {
                        throw new Error('module not found');
                    }
                    return new Module(res.set[0]);
                });
            });
        });
    }
    /**
     * Creates new Namespace object
     *
     * @example
     * // Creates enabled (!) namespace with slug & name
     * Compose.saveNamespace(Compose.makeNamespace({
     *   slug: 'my-namespace',
     *   name: 'My Namespace',
     * }))
     *
     * @param namespace
     * @param namespace, defaults to current $namespace
     */
    makeNamespace() {
        return __awaiter(this, arguments, void 0, function* (namespace = {}) {
            return new Namespace(Object.assign({ name: namespace.name || namespace.slug, meta: {}, enabled: true }, namespace));
        });
    }
    /**
     * Creates/updates Namespace
     *
     * @example
     * Compose.saveNamespace(myNamespace)
     *
     * @param namespace
     */
    saveNamespace(namespace) {
        return __awaiter(this, void 0, void 0, function* () {
            return Promise.resolve(namespace).then(namespace => {
                if (!(namespace instanceof Namespace)) {
                    throw Error('expecting Namespace type');
                }
                if (isFresh(namespace.namespaceID)) {
                    return this.ComposeAPI.namespaceCreate(kv(namespace)).then(n => new Namespace(n));
                }
                else {
                    return this.ComposeAPI.namespaceUpdate(kv(namespace)).then(n => new Namespace(n));
                }
            });
        });
    }
    /**
     * Searches for namespaces
     *
     * @private
     * @param filter
     */
    findNamespaces() {
        return __awaiter(this, arguments, void 0, function* (filter = '') {
            if (typeof filter === 'string') {
                filter = { query: filter };
            }
            return this.ComposeAPI.namespaceList(Object.assign({}, filter)).then(res => {
                // Casting all we got to to Namespace
                res.set = res.set.map(m => new Namespace(m));
                return res;
            });
        });
    }
    /**
     * Finds namespace by ID
     *
     * @example
     * // Explicitly load namespace and do something with it
     * Compose.findNamespaceByID('2039248239042').then(myNamespace => {
     *   // do something with myNamespace
     *   return Compose.findModules(myNamespace)
     * }).then(modules => {})
     *
     * // even shorter
     * Compose.findModules('2039248239042').then(....)
     *
     * @param ns - accepts Namespace, namespaceID (when string string) or Record
     */
    findNamespaceByID() {
        return __awaiter(this, arguments, void 0, function* (ns = this.$namespace) {
            const namespaceID = extractID(ns, 'namespaceID');
            return this.ComposeAPI.namespaceRead({ namespaceID }).then(m => new Namespace(m));
        });
    }
    /**
     * Finds namespace by name
     *
     * @example
     * // Explicitly load namespace and do something with it
     * Compose.findNamespaceBySlug('SomeNamespace').then(myNamespace => {
     *   // do something with myNamespace
     *   return Compose.findModules(myNamespace)
     * }).then(modules => {})
     *
     * // even shorter
     * Compose.findModules('SomeNamespace').then(....)
     *
     * @param slug - name of the namespace
     */
    findNamespaceBySlug(slug) {
        return __awaiter(this, void 0, void 0, function* () {
            return this.ComposeAPI.namespaceList({ slug }).then(res => {
                if (!Array.isArray(res.set) || res.set.length === 0) {
                    throw new Error('namespace not found');
                }
                return new Namespace(res.set[0]);
            });
        });
    }
    /**
     * Sends a simple email message
     *
     * @example
     * Compose.sendMail('some-address@domain.tld', 'subject...', { html: 'Hello!' })
     *
     * @param to - Recipient(s)
     * @param subject - Mail subject
     * @param body
     * @property {string} body.html - HTML body to be sent
     * @param Any additional addresses we want this to be sent to (carbon-copy)
     */
    sendMail(to_1, subject_1) {
        return __awaiter(this, arguments, void 0, function* (to, subject, { html = '' } = {}, { cc = [] } = {}) {
            if (!to) {
                throw Error('expecting to email address');
            }
            if (!subject) {
                throw Error('expecting subject');
            }
            if (!html) {
                throw Error('expecting HTML body');
            }
            return this.ComposeAPI.notificationEmailSend({
                to: Array.isArray(to) ? to : [to],
                cc: Array.isArray(cc) ? cc : [cc],
                subject,
                content: { html },
            });
        });
    }
    /**
     * Generates HTML with all records fields and sends it to
     *
     * @example
     * // Simplified version, sends current email with generated
     * // subject (<module name> + 'record' +  'update'/'created')
     * Compose.sendRecordToMail('example@domain.tld')
     *
     * // Complex notification with custom subject, header and footer text and custom record
     * Compose.sendRecordToMail(
     *   'asignee@domain.tld',
     *   'New lead assigned to you',
     *   {
     *      header: '<h1>New lead was created and assigned to you</h1>',
     *      footer: 'Review and confirm',
     *      cc: [ 'sales@domain.tld' ],
     *      fields: ['name', 'country', 'amount'],
     *   },
     *   newLead
     * )
     *
     * @param to - Recipient(s)
     * @param subject - Mail subject
     * @param options - Various options for body & email
     * @property {string} options.header - Text (HTML) before the record table
     * @property {string} options.footer - Text (HTML) after the record table
     * @property {string} options.style - Custom CSS styles for the email
     * @param options.fields - List of record fields we want to output
     * @param options.header - Additional mail headers (cc)
     * @param record - record to be converted (or leave for the current $record)
     */
    sendRecordToMail(to_1) {
        return __awaiter(this, arguments, void 0, function* (to, subject = '', _a = {}, record) {
            // Wait for the record if we got a promise
            var { header = '', footer = '', style = emailStyle, fields = null } = _a, mailHeader = __rest(_a, ["header", "footer", "style", "fields"]);
            if (record === void 0) { record = this.$record; }
            record = yield record;
            if (!record) {
                throw Error('record undefined');
            }
            const wb = '<div style="width: 800px; margin: 20px auto;">';
            const wa = '</div>';
            header = `${wb}${header}${wa}`;
            footer = `${wb}${footer}${wa}`;
            style = `<style type="text/css">${style}</style>`;
            const html = style + header + this.recordToHTML(fields, record) + footer;
            if (!subject) {
                subject = record.module.name + ' ';
                subject += record.updatedAt ? 'record updated' : 'record created';
            }
            return this.sendMail(to, subject, { html }, Object.assign({}, mailHeader));
        });
    }
    /**
     * Walks over white listed fields.
     *
     * @param fwl - field white list; if not defined, all fields are used
     * @param record - record to be walked over
     * @param formatter
     *
     * @private
     */
    walkFields(fwl, record, formatter) {
        if (!formatter) {
            throw new Error('formatter.undefined');
        }
        if (isRecord(fwl)) {
            record = fwl;
            fwl = undefined;
        }
        if (Array.isArray(fwl) && fwl.length === 0) {
            fwl = null;
        }
        return record.module.fields
            .filter(f => !fwl || fwl.indexOf(f.name) > -1)
            .map(formatter);
    }
    /**
     * Sends a simple record report as HTML
     *
     * @example
     * // generates report for current $record with all fields
     * let report = recordToHTML()
     *
     * // generates report for current $record from a list of fields
     * let report = recordToHTML(['fieldA', 'fieldB', 'fieldC'])
     *
     *
     * @param fwl - field white list (or leave empty/null/false for all fields)
     * @param record - record to be converted (or leave for the current $record)
     */
    recordToHTML(fwl = null, record = this.$record) {
        if (!record) {
            throw Error('record undefined');
        }
        const rows = this
            .walkFields(fwl, record, (f) => {
            const { name, label } = f;
            const v = record.values[name];
            return `<tr><td>${label || name}</td><td>${(Array.isArray(v) ? v : [v]).join(', ') || '&nbsp;'}</td></tr>`;
        })
            .join('');
        return `<table width="800" cellspacing="0" cellpadding="0" border="0">${rows}</table>`;
    }
    /**
     * Represents a given record as plain text
     *
     * @example
     * // generates report for current $record with all fields
     * let report = recordToPlainText()
     *
     * // generates report for current $record from a list of fields
     * let report = recordToPlainText(['fieldA', 'fieldB', 'fieldC'])
     *
     * @param fwl - field white list (or leave empty/null/false for all fields)
     * @param record - record to be converted (or leave for the current $record)
     */
    recordToPlainText(fwl = null, record = this.$record) {
        if (!record) {
            throw Error('record undefined');
        }
        return this
            .walkFields(fwl, record, f => {
            const { name, label } = f;
            const v = record.values[name];
            return `${label || name}:\n${(Array.isArray(v) ? v : [v]).join(', ') || '/'}\n\n`;
        })
            .join('')
            .trim();
    }
    /**
     * Scans all given arguments and returns first one that resembles something like a valid module, its name or ID
     *
     * @private
     */
    resolveModule(...args) {
        return __awaiter(this, void 0, void 0, function* () {
            const strResolve = (module) => __awaiter(this, void 0, void 0, function* () {
                return this.findModuleByHandle(module)
                    .then(m => {
                    if (!m) {
                        throw new Error('ModuleNotFound');
                    }
                    return m;
                })
                    .catch(() => this.findModuleByName(module));
            });
            for (let module of args) {
                if (!module) {
                    continue;
                }
                if (typeof module === 'string') {
                    if (IsCortezaID(module)) {
                        // Looks like an ID
                        return this.findModuleByID(module).catch((err = {}) => {
                            if (err.message && err.message.indexOf('ModuleNotFound') >= 0) {
                                // Not found, let's try if we can find it by slug
                                return strResolve(module);
                            }
                            return Promise.reject(err);
                        });
                    }
                    // Assume name
                    return strResolve(module);
                }
                if (typeof module !== 'object') {
                    continue;
                }
                // resolve whatever object we have (maybe it's a promise?)
                // and wait for results
                module = yield module;
                if (isRecord(module)) {
                    const m = module;
                    return this.resolveModule(m.module, m.moduleID);
                }
                if (IsOf(module, 'set', 'filter')) {
                    // We got a result set with modules
                    module = module.set;
                }
                if (Array.isArray(module)) {
                    // We got array of modules
                    if (module.length === 0) {
                        // Empty array
                        continue;
                    }
                    else {
                        // Use first module from the list
                        module = module.shift();
                    }
                }
                if (!isModule(module)) {
                    // not module? is it an object with moduleID & namespaceID?
                    if (module.moduleID === undefined || module.namespaceID === undefined) {
                        break;
                    }
                    return Promise.resolve(new Module(module));
                }
                return Promise.resolve(module);
            }
            return Promise.reject(Error('unexpected input type for module resolver'));
        });
    }
    /**
     * Scans all given arguments and returns first one that resembles something like a valid namespace, its slug or ID
     *
     * @private
     */
    resolveNamespace(...args) {
        return __awaiter(this, void 0, void 0, function* () {
            for (let ns of args) {
                if (!ns) {
                    continue;
                }
                if (typeof ns === 'string') {
                    if (IsCortezaID(ns)) {
                        // Looks like an ID
                        return this.findNamespaceByID(ns).catch((err = {}) => {
                            if (err.message && err.message.indexOf('NamespaceNotFound') >= 0) {
                                // Not found, let's try if we can find it by slug
                                return this.findNamespaceBySlug(ns);
                            }
                            return Promise.reject(err);
                        });
                    }
                    // Assume namespace slug
                    return this.findNamespaceBySlug(ns);
                }
                if (typeof ns !== 'object') {
                    continue;
                }
                // resolve whatever object we have (maybe it's a promise?)
                // and wait for results
                ns = yield ns;
                if (isRecord(ns)) {
                    const n = ns;
                    return this.resolveNamespace(n.namespaceID);
                }
                if (ns.set && ns.filter) {
                    // We got a result set with modules
                    ns = ns.set;
                }
                if (Array.isArray(ns)) {
                    // We got array of modules
                    if (ns.length === 0) {
                        // Empty array
                        continue;
                    }
                    else {
                        // Use first module from the list
                        ns = ns.shift();
                    }
                }
                if (!(ns instanceof Namespace)) {
                    // not Namespace? is it an object with namespaceID?
                    if (ns.namespaceID === undefined) {
                        break;
                    }
                    return Promise.resolve(new Namespace(ns));
                }
                return Promise.resolve(ns);
            }
            return Promise.reject(Error('unexpected input type for namespace resolver'));
        });
    }
    /**
     * Allows access for the given role for the given Compose resource
     *
     * @example
     * // Allows users with `someRole` to access the newly created namespace
     * await Compose.allow({
     *    role: someRole,
     *    resource: newNamespace,
     *    operation: 'read',
     * })
     */
    allow(...pr) {
        return __awaiter(this, void 0, void 0, function* () {
            const rr = pr.map(p => ({
                role: p.role,
                resource: p.resource,
                operation: p.operation,
                access: 'allow',
            }));
            return genericPermissionUpdater(this.ComposeAPI, rr);
        });
    }
    /**
     * Denies access for the given role for the given Compose resource
     *
     * @example
     * // Denies users with `someRole` from accessing the newly created namespace
     * await Compose.deny({
     *    role: someRole,
     *    resource: newNamespace,
     *    operation: 'read',
     * })
     */
    deny(...pr) {
        return __awaiter(this, void 0, void 0, function* () {
            const rr = pr.map(p => ({
                role: p.role,
                resource: p.resource,
                operation: p.operation,
                access: 'deny',
            }));
            return genericPermissionUpdater(this.ComposeAPI, rr);
        });
    }
    /**
     * Inherits access for the given role for the given Compose resource
     *
     * @example
     * // Uses inherited permissions for the `sameRole` for the newly created namespace
     * await Compose.inherit({
     *    role: someRole,
     *    resource: newNamespace,
     *    operation: 'read',
     * })
     */
    inherit(...pr) {
        return __awaiter(this, void 0, void 0, function* () {
            const rr = pr.map(p => ({
                role: p.role,
                resource: p.resource,
                operation: p.operation,
                access: 'inherit',
            }));
            return genericPermissionUpdater(this.ComposeAPI, rr);
        });
    }
}

/**
 * Generic type caster
 *
 * Takes argument (ref to class) and returns a function that will initialize class of that type
 */
function GenericCaster(C) {
    return function (val) {
        if (!val || typeof val !== 'object') {
            return undefined;
        }
        return new C(val);
    };
}
/**
 * Generic type caster with Object.freeze
 *
 * Takes argument (ref to class) and returns a function that will initialize class of that type
 */
function GenericCasterFreezer(C) {
    return function (val) {
        if (!val || typeof val !== 'object') {
            return undefined;
        }
        return Object.freeze(new C(val));
    };
}

/**
 * Record type caster
 *
 * Record arg is a bit special, it takes 2 params (record itself + record's module)
 */
function recordCaster(val) {
    if (val) {
        try {
            return new Record(this.$module, val);
        }
        catch (e) {
            console.error(e);
        }
    }
    return undefined;
}
function recordCasterFreezer(val) {
    if (val) {
        try {
            return Object.freeze(new Record(this.$module, val));
        }
        catch (e) {
            console.error(e);
        }
    }
    return undefined;
}
/**
 * CortezaTypes map helps ExecArgs class with translation of (special) arguments
 * to their respected types
 *
 * There's noe need to set/define casters for old* arguments,
 * It's auto-magically done by Args class
 */
const CortezaTypes = new Map();
CortezaTypes.set('authUser', GenericCasterFreezer(User));
CortezaTypes.set('invoker', GenericCasterFreezer(User));
CortezaTypes.set('module', GenericCaster(Module));
CortezaTypes.set('oldModule', GenericCasterFreezer(Module));
CortezaTypes.set('page', GenericCaster(Page));
CortezaTypes.set('oldPage', GenericCasterFreezer(Page));
CortezaTypes.set('namespace', GenericCaster(Namespace));
CortezaTypes.set('oldNamespace', GenericCasterFreezer(Namespace));
CortezaTypes.set('application', GenericCaster(Application));
CortezaTypes.set('oldApplication', GenericCasterFreezer(Application));
CortezaTypes.set('user', GenericCaster(User));
CortezaTypes.set('oldUser', GenericCasterFreezer(User));
CortezaTypes.set('role', GenericCaster(Role));
CortezaTypes.set('oldRole', GenericCasterFreezer(Role));
CortezaTypes.set('record', recordCaster);
CortezaTypes.set('oldRecord', recordCasterFreezer);
CortezaTypes.set('request', GenericCasterFreezer(SinkRequest));
CortezaTypes.set('response', GenericCaster(SinkResponse));

/**
 * Handles arguments, passed to the script
 *
 * By convention variables holding "current" resources are prefixed with dollar ($) sign.
 * For example, before/after triggers for record will call registered scripts with $record, $module
 * and $namespace, holding current record, it's module and namespace.
 *
 * All these variables are casted (if passed as an argument) to proper types ($record => Record, $module => Module, ...)
 */
class Args {
    constructor(args, caster = CortezaTypes) {
        const cachedArgs = {};
        for (const arg in args) {
            if (caster && caster.has(arg)) {
                const cast = caster.get(arg);
                if (cast) {
                    Object.defineProperty(this, `$${arg}`, {
                        get: () => {
                            if (!cachedArgs[arg]) {
                                cachedArgs[arg] = cast.call(this, args[arg]);
                            }
                            return cachedArgs[arg];
                        },
                        configurable: false,
                        enumerable: true,
                    });
                }
                Object.defineProperty(this, `raw${arg.substring(0, 1).toUpperCase()}${arg.substring(1)}`, {
                    value: args[arg],
                    writable: false,
                    configurable: false,
                    enumerable: true,
                });
            }
            else {
                Object.defineProperty(this, arg, {
                    value: args[arg],
                    writable: false,
                    configurable: false,
                    enumerable: true,
                });
            }
        }
    }
}
/**
 * Handles arguments, passed to the script but preserves references to the original objects
 *
 * By convention variables holding "current" resources are prefixed with dollar ($) sign.
 * For example, before/after triggers for record will call registered scripts with $record, $module
 * and $namespace, holding current record, it's module and namespace.
 *
 * These variables are not additionally casted, since in order to preserve references they should
 * already be in the correct type.
 */
class ArgsProxy {
    constructor(args, caster = CortezaTypes) {
        for (const arg in args) {
            // For consistency only prefix args with & and raw that have a defined caster
            if (caster && caster.has(arg)) {
                Object.defineProperty(this, `$${arg}`, {
                    get: () => args[arg],
                    configurable: false,
                    enumerable: true,
                });
                Object.defineProperty(this, `raw${arg.substring(0, 1).toUpperCase()}${arg.substring(1)}`, {
                    value: args[arg],
                    writable: false,
                    configurable: false,
                    enumerable: true,
                });
            }
            else {
                Object.defineProperty(this, arg, {
                    value: args[arg],
                    writable: false,
                    configurable: false,
                    enumerable: true,
                });
            }
        }
    }
}

/* eslint-disable padded-blocks */
function stdResolve$4(response) {
    if (response.data.error) {
        return Promise.reject(response.data.error);
    }
    else {
        return response.data.response;
    }
}
class System {
    constructor({ baseURL, headers, accessTokenFn }) {
        this.headers = {};
        this.baseURL = baseURL;
        this.accessTokenFn = accessTokenFn;
        this.headers = {
            /**
             * All we send is JSON
             */
            'Content-Type': 'application/json',
        };
        this.setHeaders(headers);
    }
    setAccessTokenFn(fn) {
        this.accessTokenFn = fn;
        return this;
    }
    setHeaders(headers) {
        if (typeof headers === 'object') {
            this.headers = headers;
        }
        return this;
    }
    setHeader(name, value) {
        if (value === undefined) {
            delete this.headers[name];
        }
        else {
            this.headers[name] = value;
        }
        return this;
    }
    api() {
        const headers = Object.assign({}, this.headers);
        const accessToken = this.accessTokenFn ? this.accessTokenFn() : undefined;
        if (accessToken) {
            headers.Authorization = 'Bearer ' + accessToken;
        }
        return axios.create({
            withCredentials: true,
            baseURL: this.baseURL,
            headers,
        });
    }
    // Impersonate a user
    authImpersonate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.authImpersonateEndpoint() });
            cfg.data = {
                userID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    authImpersonateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.authImpersonate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    authImpersonateEndpoint() {
        return '/auth/impersonate';
    }
    // List clients
    authClientList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { handle, deleted, labels, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.authClientListEndpoint() });
            cfg.params = {
                handle,
                deleted,
                labels,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    authClientListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.authClientList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    authClientListEndpoint() {
        return '/auth/clients/';
    }
    // Create client
    authClientCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { handle, meta, validGrant, redirectURI, scope, trusted, enabled, validFrom, expiresAt, security, labels, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.authClientCreateEndpoint() });
            cfg.data = {
                handle,
                meta,
                validGrant,
                redirectURI,
                scope,
                trusted,
                enabled,
                validFrom,
                expiresAt,
                security,
                labels,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    authClientCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.authClientCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    authClientCreateEndpoint() {
        return '/auth/clients/';
    }
    // Update user details
    authClientUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { clientID, handle, meta, validGrant, redirectURI, scope, trusted, enabled, validFrom, expiresAt, security, labels, updatedAt, } = a || {};
            if (!clientID) {
                throw Error('field clientID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.authClientUpdateEndpoint({
                    clientID,
                }) });
            cfg.data = {
                handle,
                meta,
                validGrant,
                redirectURI,
                scope,
                trusted,
                enabled,
                validFrom,
                expiresAt,
                security,
                labels,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    authClientUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.authClientUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    authClientUpdateEndpoint(a) {
        const { clientID, } = a || {};
        return `/auth/clients/${clientID}`;
    }
    // Read client details
    authClientRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { clientID, } = a || {};
            if (!clientID) {
                throw Error('field clientID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.authClientReadEndpoint({
                    clientID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    authClientReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.authClientRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    authClientReadEndpoint(a) {
        const { clientID, } = a || {};
        return `/auth/clients/${clientID}`;
    }
    // Remove client
    authClientDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { clientID, } = a || {};
            if (!clientID) {
                throw Error('field clientID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.authClientDeleteEndpoint({
                    clientID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    authClientDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.authClientDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    authClientDeleteEndpoint(a) {
        const { clientID, } = a || {};
        return `/auth/clients/${clientID}`;
    }
    // Undelete client
    authClientUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { clientID, } = a || {};
            if (!clientID) {
                throw Error('field clientID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.authClientUndeleteEndpoint({
                    clientID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    authClientUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.authClientUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    authClientUndeleteEndpoint(a) {
        const { clientID, } = a || {};
        return `/auth/clients/${clientID}/undelete`;
    }
    // Regenerate client&#x27;s secret
    authClientRegenerateSecret(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { clientID, } = a || {};
            if (!clientID) {
                throw Error('field clientID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.authClientRegenerateSecretEndpoint({
                    clientID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    authClientRegenerateSecretCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.authClientRegenerateSecret(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    authClientRegenerateSecretEndpoint(a) {
        const { clientID, } = a || {};
        return `/auth/clients/${clientID}/secret`;
    }
    // Exposes client&#x27;s secret
    authClientExposeSecret(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { clientID, } = a || {};
            if (!clientID) {
                throw Error('field clientID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.authClientExposeSecretEndpoint({
                    clientID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    authClientExposeSecretCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.authClientExposeSecret(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    authClientExposeSecretEndpoint(a) {
        const { clientID, } = a || {};
        return `/auth/clients/${clientID}/secret`;
    }
    // Evaluate expressions
    expressionEvaluate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { variables, expressions, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.expressionEvaluateEndpoint() });
            cfg.data = {
                variables,
                expressions,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    expressionEvaluateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.expressionEvaluate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    expressionEvaluateEndpoint() {
        return '/expressions/evaluate';
    }
    // List settings
    settingsList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { prefix, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.settingsListEndpoint() });
            cfg.params = {
                prefix,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    settingsListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.settingsList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    settingsListEndpoint() {
        return '/settings/';
    }
    // Update settings
    settingsUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { values, } = a || {};
            if (!values) {
                throw Error('field values is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.settingsUpdateEndpoint() });
            cfg.data = {
                values,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    settingsUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.settingsUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    settingsUpdateEndpoint() {
        return '/settings/';
    }
    // Get a value for a key
    settingsGet(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { key, ownerID, } = a || {};
            if (!key) {
                throw Error('field key is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.settingsGetEndpoint({
                    key,
                }) });
            cfg.params = {
                ownerID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    settingsGetCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.settingsGet(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    settingsGetEndpoint(a) {
        const { key, } = a || {};
        return `/settings/${key}`;
    }
    // Set value for specific setting
    settingsSet(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { key, upload, ownerID, } = a || {};
            if (!key) {
                throw Error('field key is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.settingsSetEndpoint({
                    key,
                }) });
            cfg.data = {
                upload,
                ownerID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    settingsSetCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.settingsSet(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    settingsSetEndpoint(a) {
        const { key, } = a || {};
        return `/settings/${key}`;
    }
    // Current compose settings
    settingsCurrent() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.settingsCurrentEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    settingsCurrentCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.settingsCurrent(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    settingsCurrentEndpoint() {
        return '/settings/current';
    }
    // List roles
    roleList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { query, memberID, userGroupID, roleID, deleted, archived, labels, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.roleListEndpoint() });
            cfg.params = {
                query,
                memberID,
                userGroupID,
                roleID,
                deleted,
                archived,
                labels,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleListEndpoint() {
        return '/roles/';
    }
    // Update role details
    roleCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { name, handle, members, meta, labels, } = a || {};
            if (!name) {
                throw Error('field name is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleCreateEndpoint() });
            cfg.data = {
                name,
                handle,
                members,
                meta,
                labels,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleCreateEndpoint() {
        return '/roles/';
    }
    // Update role details
    roleUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, name, handle, members, meta, labels, updatedAt, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.roleUpdateEndpoint({
                    roleID,
                }) });
            cfg.data = {
                name,
                handle,
                members,
                meta,
                labels,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleUpdateEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}`;
    }
    // Read role details and memberships
    roleRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.roleReadEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleReadEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}`;
    }
    // Remove role
    roleDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.roleDeleteEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleDeleteEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}`;
    }
    // Archive role
    roleArchive(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleArchiveEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleArchiveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleArchive(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleArchiveEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}/archive`;
    }
    // Unarchive role
    roleUnarchive(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleUnarchiveEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleUnarchiveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleUnarchive(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleUnarchiveEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}/unarchive`;
    }
    // Undelete role
    roleUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleUndeleteEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleUndeleteEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}/undelete`;
    }
    // Move role to different organisation
    roleMove(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, organisationID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!organisationID) {
                throw Error('field organisationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleMoveEndpoint({
                    roleID,
                }) });
            cfg.data = {
                organisationID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleMoveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleMove(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleMoveEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}/move`;
    }
    // Merge one role into another
    roleMerge(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, destination, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!destination) {
                throw Error('field destination is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleMergeEndpoint({
                    roleID,
                }) });
            cfg.data = {
                destination,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleMergeCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleMerge(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleMergeEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}/merge`;
    }
    // Returns all role members
    roleMemberList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.roleMemberListEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleMemberListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleMemberList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleMemberListEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}/members`;
    }
    // Add user group to a user group
    roleMemberAddGroup(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, userGroupID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleMemberAddGroupEndpoint({
                    roleID, userGroupID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleMemberAddGroupCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleMemberAddGroup(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleMemberAddGroupEndpoint(a) {
        const { roleID, userGroupID, } = a || {};
        return `/roles/${roleID}/member-g/${userGroupID}`;
    }
    // Add member to a role
    roleMemberAdd(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, userID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleMemberAddEndpoint({
                    roleID, userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleMemberAddCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleMemberAdd(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleMemberAddEndpoint(a) {
        const { roleID, userID, } = a || {};
        return `/roles/${roleID}/member/${userID}`;
    }
    // Remove member from a role
    roleMemberRemove(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, userID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.roleMemberRemoveEndpoint({
                    roleID, userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleMemberRemoveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleMemberRemove(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleMemberRemoveEndpoint(a) {
        const { roleID, userID, } = a || {};
        return `/roles/${roleID}/member/${userID}`;
    }
    // Remove user group from a role
    roleMemberRemoveGroup(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, userGroupID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.roleMemberRemoveGroupEndpoint({
                    roleID, userGroupID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleMemberRemoveGroupCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleMemberRemoveGroup(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleMemberRemoveGroupEndpoint(a) {
        const { roleID, userGroupID, } = a || {};
        return `/roles/${roleID}/member-g/${userGroupID}`;
    }
    // Fire system:role trigger
    roleTriggerScript(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, script, args, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!script) {
                throw Error('field script is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleTriggerScriptEndpoint({
                    roleID,
                }) });
            cfg.data = {
                script,
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleTriggerScriptCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleTriggerScript(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleTriggerScriptEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}/trigger`;
    }
    // Clone permission settings to a role
    roleCloneRules(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, cloneToRoleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!cloneToRoleID) {
                throw Error('field cloneToRoleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.roleCloneRulesEndpoint({
                    roleID,
                }) });
            cfg.params = {
                cloneToRoleID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    roleCloneRulesCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.roleCloneRules(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    roleCloneRulesEndpoint(a) {
        const { roleID, } = a || {};
        return `/roles/${roleID}/rules/clone`;
    }
    // List user groups
    userGroupList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { query, memberID, userGroupID, deleted, archived, labels, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.userGroupListEndpoint() });
            cfg.params = {
                query,
                memberID,
                userGroupID,
                deleted,
                archived,
                labels,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userGroupListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userGroupList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userGroupListEndpoint() {
        return '/user-groups/';
    }
    // Update user groups details
    userGroupCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { handle, members, config, meta, labels, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userGroupCreateEndpoint() });
            cfg.data = {
                handle,
                members,
                config,
                meta,
                labels,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userGroupCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userGroupCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userGroupCreateEndpoint() {
        return '/user-groups/';
    }
    // Update user group details
    userGroupUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userGroupID, handle, members, config, meta, labels, updatedAt, } = a || {};
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.userGroupUpdateEndpoint({
                    userGroupID,
                }) });
            cfg.data = {
                handle,
                members,
                config,
                meta,
                labels,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userGroupUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userGroupUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userGroupUpdateEndpoint(a) {
        const { userGroupID, } = a || {};
        return `/user-groups/${userGroupID}`;
    }
    // Read user group details and memberships
    userGroupRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userGroupID, } = a || {};
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.userGroupReadEndpoint({
                    userGroupID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userGroupReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userGroupRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userGroupReadEndpoint(a) {
        const { userGroupID, } = a || {};
        return `/user-groups/${userGroupID}`;
    }
    // Remove user group
    userGroupDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userGroupID, } = a || {};
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.userGroupDeleteEndpoint({
                    userGroupID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userGroupDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userGroupDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userGroupDeleteEndpoint(a) {
        const { userGroupID, } = a || {};
        return `/user-groups/${userGroupID}`;
    }
    // Undelete user group
    userGroupUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userGroupID, } = a || {};
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userGroupUndeleteEndpoint({
                    userGroupID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userGroupUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userGroupUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userGroupUndeleteEndpoint(a) {
        const { userGroupID, } = a || {};
        return `/user-groups/${userGroupID}/undelete`;
    }
    // Returns all user group members
    userGroupMemberList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userGroupID, } = a || {};
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.userGroupMemberListEndpoint({
                    userGroupID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userGroupMemberListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userGroupMemberList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userGroupMemberListEndpoint(a) {
        const { userGroupID, } = a || {};
        return `/user-groups/${userGroupID}/members`;
    }
    // Add member to a user group
    userGroupMemberAdd(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userGroupID, userID, } = a || {};
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userGroupMemberAddEndpoint({
                    userGroupID, userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userGroupMemberAddCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userGroupMemberAdd(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userGroupMemberAddEndpoint(a) {
        const { userGroupID, userID, } = a || {};
        return `/user-groups/${userGroupID}/member/${userID}`;
    }
    // Search users (Directory)
    userList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, roleID, query, username, email, handle, kind, incDeleted, incSuspended, deleted, suspended, labels, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.userListEndpoint() });
            cfg.params = {
                userID,
                roleID,
                query,
                username,
                email,
                handle,
                kind,
                incDeleted,
                incSuspended,
                deleted,
                suspended,
                labels,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userListEndpoint() {
        return '/users/';
    }
    // Create user
    userCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { email, name, handle, userGroupID, kind, labels, meta, } = a || {};
            if (!email) {
                throw Error('field email is empty');
            }
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userCreateEndpoint() });
            cfg.data = {
                email,
                name,
                handle,
                userGroupID,
                kind,
                labels,
                meta,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userCreateEndpoint() {
        return '/users/';
    }
    // Update user details
    userUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, email, name, handle, userGroupID, kind, labels, meta, updatedAt, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            if (!email) {
                throw Error('field email is empty');
            }
            if (!userGroupID) {
                throw Error('field userGroupID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.userUpdateEndpoint({
                    userID,
                }) });
            cfg.data = {
                email,
                name,
                handle,
                userGroupID,
                kind,
                labels,
                meta,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userUpdateEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}`;
    }
    // Patch user (experimental)
    userPartialUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.userPartialUpdateEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userPartialUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userPartialUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userPartialUpdateEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}`;
    }
    // Read user details
    userRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.userReadEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userReadEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}`;
    }
    // Remove user
    userDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.userDeleteEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userDeleteEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}`;
    }
    // Suspend user
    userSuspend(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userSuspendEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userSuspendCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userSuspend(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userSuspendEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/suspend`;
    }
    // Unsuspend user
    userUnsuspend(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userUnsuspendEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userUnsuspendCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userUnsuspend(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userUnsuspendEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/unsuspend`;
    }
    // Undelete user
    userUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userUndeleteEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userUndeleteEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/undelete`;
    }
    // Set&#x27;s or changes user&#x27;s password
    userSetPassword(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, password, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userSetPasswordEndpoint({
                    userID,
                }) });
            cfg.data = {
                password,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userSetPasswordCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userSetPassword(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userSetPasswordEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/password`;
    }
    // Add member to a role
    userMembershipList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.userMembershipListEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userMembershipListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userMembershipList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userMembershipListEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/membership`;
    }
    // Add role to a user
    userMembershipAdd(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, userID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userMembershipAddEndpoint({
                    roleID, userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userMembershipAddCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userMembershipAdd(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userMembershipAddEndpoint(a) {
        const { roleID, userID, } = a || {};
        return `/users/${userID}/membership/${roleID}`;
    }
    // Remove role from a user
    userMembershipRemove(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, userID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.userMembershipRemoveEndpoint({
                    roleID, userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userMembershipRemoveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userMembershipRemove(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userMembershipRemoveEndpoint(a) {
        const { roleID, userID, } = a || {};
        return `/users/${userID}/membership/${roleID}`;
    }
    // Fire system:user trigger
    userTriggerScript(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, script, args, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            if (!script) {
                throw Error('field script is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userTriggerScriptEndpoint({
                    userID,
                }) });
            cfg.data = {
                script,
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userTriggerScriptCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userTriggerScript(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userTriggerScriptEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/trigger`;
    }
    // Remove all auth sessions of user
    userSessionsRemove(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.userSessionsRemoveEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userSessionsRemoveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userSessionsRemove(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userSessionsRemoveEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/sessions`;
    }
    // List user&#x27;s credentials
    userListCredentials(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.userListCredentialsEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userListCredentialsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userListCredentials(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userListCredentialsEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/credentials`;
    }
    // List user&#x27;s credentials
    userDeleteCredentials(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, credentialsID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            if (!credentialsID) {
                throw Error('field credentialsID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.userDeleteCredentialsEndpoint({
                    userID, credentialsID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userDeleteCredentialsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userDeleteCredentials(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userDeleteCredentialsEndpoint(a) {
        const { userID, credentialsID, } = a || {};
        return `/users/${userID}/credentials/${credentialsID}`;
    }
    // User&#x27;s profile avatar
    userProfileAvatar(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, upload, width, height, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userProfileAvatarEndpoint({
                    userID,
                }) });
            cfg.data = {
                upload,
                width,
                height,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userProfileAvatarCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userProfileAvatar(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userProfileAvatarEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/avatar`;
    }
    // User profile avatar initial
    userProfileAvatarInitial(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, avatarColor, avatarBgColor, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userProfileAvatarInitialEndpoint({
                    userID,
                }) });
            cfg.data = {
                avatarColor,
                avatarBgColor,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userProfileAvatarInitialCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userProfileAvatarInitial(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userProfileAvatarInitialEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/avatar-initial`;
    }
    // delete user&#x27;s profile avatar
    userDeleteAvatar(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { userID, } = a || {};
            if (!userID) {
                throw Error('field userID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.userDeleteAvatarEndpoint({
                    userID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userDeleteAvatarCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userDeleteAvatar(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userDeleteAvatarEndpoint(a) {
        const { userID, } = a || {};
        return `/users/${userID}/avatar`;
    }
    // Export users
    userExport(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { filename, inclRoleMembership, inclRoles, } = a || {};
            if (!filename) {
                throw Error('field filename is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.userExportEndpoint({
                    filename,
                }) });
            cfg.params = {
                inclRoleMembership,
                inclRoles,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userExportCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userExport(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userExportEndpoint(a) {
        const { filename, } = a || {};
        return `/users/export/${filename}.zip`;
    }
    // Import users
    userImport(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { upload, } = a || {};
            if (!upload) {
                throw Error('field upload is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.userImportEndpoint() });
            cfg.data = {
                upload,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    userImportCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.userImport(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    userImportEndpoint() {
        return '/users/import';
    }
    // Search drivers
    dalDriverList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dalDriverListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalDriverListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalDriverList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalDriverListEndpoint() {
        return '/dal/drivers/';
    }
    // Search sensitivity levels
    dalSensitivityLevelList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sensitivityLevelID, deleted, incTotal, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dalSensitivityLevelListEndpoint() });
            cfg.params = {
                sensitivityLevelID,
                deleted,
                incTotal,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSensitivityLevelListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSensitivityLevelList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSensitivityLevelListEndpoint() {
        return '/dal/sensitivity-levels/';
    }
    // Create sensitivity level
    dalSensitivityLevelCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { handle, level, meta, } = a || {};
            if (!level) {
                throw Error('field level is empty');
            }
            if (!meta) {
                throw Error('field meta is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.dalSensitivityLevelCreateEndpoint() });
            cfg.data = {
                handle,
                level,
                meta,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSensitivityLevelCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSensitivityLevelCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSensitivityLevelCreateEndpoint() {
        return '/dal/sensitivity-levels/';
    }
    // Update sensitivity details
    dalSensitivityLevelUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sensitivityLevelID, handle, level, meta, updatedAt, } = a || {};
            if (!sensitivityLevelID) {
                throw Error('field sensitivityLevelID is empty');
            }
            if (!level) {
                throw Error('field level is empty');
            }
            if (!meta) {
                throw Error('field meta is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.dalSensitivityLevelUpdateEndpoint({
                    sensitivityLevelID,
                }) });
            cfg.data = {
                handle,
                level,
                meta,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSensitivityLevelUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSensitivityLevelUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSensitivityLevelUpdateEndpoint(a) {
        const { sensitivityLevelID, } = a || {};
        return `/dal/sensitivity-levels/${sensitivityLevelID}`;
    }
    // Read connection details
    dalSensitivityLevelRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sensitivityLevelID, } = a || {};
            if (!sensitivityLevelID) {
                throw Error('field sensitivityLevelID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dalSensitivityLevelReadEndpoint({
                    sensitivityLevelID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSensitivityLevelReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSensitivityLevelRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSensitivityLevelReadEndpoint(a) {
        const { sensitivityLevelID, } = a || {};
        return `/dal/sensitivity-levels/${sensitivityLevelID}`;
    }
    // Remove sensitivity level
    dalSensitivityLevelDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sensitivityLevelID, } = a || {};
            if (!sensitivityLevelID) {
                throw Error('field sensitivityLevelID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.dalSensitivityLevelDeleteEndpoint({
                    sensitivityLevelID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSensitivityLevelDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSensitivityLevelDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSensitivityLevelDeleteEndpoint(a) {
        const { sensitivityLevelID, } = a || {};
        return `/dal/sensitivity-levels/${sensitivityLevelID}`;
    }
    // Undelete sensitivity level
    dalSensitivityLevelUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sensitivityLevelID, } = a || {};
            if (!sensitivityLevelID) {
                throw Error('field sensitivityLevelID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.dalSensitivityLevelUndeleteEndpoint({
                    sensitivityLevelID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSensitivityLevelUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSensitivityLevelUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSensitivityLevelUndeleteEndpoint(a) {
        const { sensitivityLevelID, } = a || {};
        return `/dal/sensitivity-levels/${sensitivityLevelID}/undelete`;
    }
    // Search schema alterations
    dalSchemaAlterationList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { alterationID, batchID, resource, resourceType, kind, deleted, completed, dismissed, incTotal, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dalSchemaAlterationListEndpoint() });
            cfg.params = {
                alterationID,
                batchID,
                resource,
                resourceType,
                kind,
                deleted,
                completed,
                dismissed,
                incTotal,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSchemaAlterationListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSchemaAlterationList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSchemaAlterationListEndpoint() {
        return '/dal/schema/alterations/';
    }
    // Read alteration details
    dalSchemaAlterationRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { alterationID, } = a || {};
            if (!alterationID) {
                throw Error('field alterationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dalSchemaAlterationReadEndpoint({
                    alterationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSchemaAlterationReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSchemaAlterationRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSchemaAlterationReadEndpoint(a) {
        const { alterationID, } = a || {};
        return `/dal/schema/alterations/${alterationID}`;
    }
    // Apply alterations
    dalSchemaAlterationApply(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { alterationID, } = a || {};
            if (!alterationID) {
                throw Error('field alterationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.dalSchemaAlterationApplyEndpoint() });
            cfg.params = {
                alterationID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSchemaAlterationApplyCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSchemaAlterationApply(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSchemaAlterationApplyEndpoint() {
        return '/dal/schema/alterations/apply';
    }
    // Dismiss alterations
    dalSchemaAlterationDismiss(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { alterationID, } = a || {};
            if (!alterationID) {
                throw Error('field alterationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.dalSchemaAlterationDismissEndpoint() });
            cfg.params = {
                alterationID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalSchemaAlterationDismissCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalSchemaAlterationDismiss(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalSchemaAlterationDismissEndpoint() {
        return '/dal/schema/alterations/dismiss';
    }
    // Search connections (Directory)
    dalConnectionList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { connectionID, handle, type, deleted, incTotal, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dalConnectionListEndpoint() });
            cfg.params = {
                connectionID,
                handle,
                type,
                deleted,
                incTotal,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalConnectionListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalConnectionList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalConnectionListEndpoint() {
        return '/dal/connections/';
    }
    // Create connection
    dalConnectionCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { handle, type, meta, config, } = a || {};
            if (!type) {
                throw Error('field type is empty');
            }
            if (!meta) {
                throw Error('field meta is empty');
            }
            if (!config) {
                throw Error('field config is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.dalConnectionCreateEndpoint() });
            cfg.data = {
                handle,
                type,
                meta,
                config,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalConnectionCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalConnectionCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalConnectionCreateEndpoint() {
        return '/dal/connections/';
    }
    // Update connection details
    dalConnectionUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { connectionID, handle, type, meta, config, updatedAt, } = a || {};
            if (!connectionID) {
                throw Error('field connectionID is empty');
            }
            if (!type) {
                throw Error('field type is empty');
            }
            if (!meta) {
                throw Error('field meta is empty');
            }
            if (!config) {
                throw Error('field config is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.dalConnectionUpdateEndpoint({
                    connectionID,
                }) });
            cfg.data = {
                handle,
                type,
                meta,
                config,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalConnectionUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalConnectionUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalConnectionUpdateEndpoint(a) {
        const { connectionID, } = a || {};
        return `/dal/connections/${connectionID}`;
    }
    // Read connection details
    dalConnectionRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { connectionID, } = a || {};
            if (!connectionID) {
                throw Error('field connectionID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dalConnectionReadEndpoint({
                    connectionID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalConnectionReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalConnectionRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalConnectionReadEndpoint(a) {
        const { connectionID, } = a || {};
        return `/dal/connections/${connectionID}`;
    }
    // Remove connection
    dalConnectionDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { connectionID, } = a || {};
            if (!connectionID) {
                throw Error('field connectionID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.dalConnectionDeleteEndpoint({
                    connectionID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalConnectionDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalConnectionDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalConnectionDeleteEndpoint(a) {
        const { connectionID, } = a || {};
        return `/dal/connections/${connectionID}`;
    }
    // Undelete connection
    dalConnectionUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { connectionID, } = a || {};
            if (!connectionID) {
                throw Error('field connectionID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.dalConnectionUndeleteEndpoint({
                    connectionID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dalConnectionUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dalConnectionUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dalConnectionUndeleteEndpoint(a) {
        const { connectionID, } = a || {};
        return `/dal/connections/${connectionID}/undelete`;
    }
    // List applications
    applicationList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { name, query, deleted, labels, flags, incFlags, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.applicationListEndpoint() });
            cfg.params = {
                name,
                query,
                deleted,
                labels,
                flags,
                incFlags,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationListEndpoint() {
        return '/application/';
    }
    // Create application
    applicationCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { name, enabled, weight, unify, config, labels, } = a || {};
            if (!name) {
                throw Error('field name is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.applicationCreateEndpoint() });
            cfg.data = {
                name,
                enabled,
                weight,
                unify,
                config,
                labels,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationCreateEndpoint() {
        return '/application/';
    }
    // Update user details
    applicationUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { applicationID, name, enabled, weight, unify, config, labels, updatedAt, } = a || {};
            if (!applicationID) {
                throw Error('field applicationID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.applicationUpdateEndpoint({
                    applicationID,
                }) });
            cfg.data = {
                name,
                enabled,
                weight,
                unify,
                config,
                labels,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationUpdateEndpoint(a) {
        const { applicationID, } = a || {};
        return `/application/${applicationID}`;
    }
    // Upload application assets
    applicationUpload(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { upload, } = a || {};
            if (!upload) {
                throw Error('field upload is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.applicationUploadEndpoint() });
            cfg.data = {
                upload,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationUploadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationUpload(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationUploadEndpoint() {
        return '/application/upload';
    }
    // Flag application
    applicationFlagCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { applicationID, flag, ownedBy, } = a || {};
            if (!applicationID) {
                throw Error('field applicationID is empty');
            }
            if (!flag) {
                throw Error('field flag is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.applicationFlagCreateEndpoint({
                    applicationID, flag, ownedBy,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationFlagCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationFlagCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationFlagCreateEndpoint(a) {
        const { applicationID, flag, ownedBy, } = a || {};
        return `/application/${applicationID}/flag/${ownedBy}/${flag}`;
    }
    // Unflag application
    applicationFlagDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { applicationID, flag, ownedBy, } = a || {};
            if (!applicationID) {
                throw Error('field applicationID is empty');
            }
            if (!flag) {
                throw Error('field flag is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.applicationFlagDeleteEndpoint({
                    applicationID, flag, ownedBy,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationFlagDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationFlagDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationFlagDeleteEndpoint(a) {
        const { applicationID, flag, ownedBy, } = a || {};
        return `/application/${applicationID}/flag/${ownedBy}/${flag}`;
    }
    // Read application details
    applicationRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { applicationID, incFlags, } = a || {};
            if (!applicationID) {
                throw Error('field applicationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.applicationReadEndpoint({
                    applicationID,
                }) });
            cfg.params = {
                incFlags,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationReadEndpoint(a) {
        const { applicationID, } = a || {};
        return `/application/${applicationID}`;
    }
    // Remove application
    applicationDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { applicationID, } = a || {};
            if (!applicationID) {
                throw Error('field applicationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.applicationDeleteEndpoint({
                    applicationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationDeleteEndpoint(a) {
        const { applicationID, } = a || {};
        return `/application/${applicationID}`;
    }
    // Undelete application
    applicationUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { applicationID, } = a || {};
            if (!applicationID) {
                throw Error('field applicationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.applicationUndeleteEndpoint({
                    applicationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationUndeleteEndpoint(a) {
        const { applicationID, } = a || {};
        return `/application/${applicationID}/undelete`;
    }
    // Fire system:application trigger
    applicationTriggerScript(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { applicationID, script, args, } = a || {};
            if (!applicationID) {
                throw Error('field applicationID is empty');
            }
            if (!script) {
                throw Error('field script is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.applicationTriggerScriptEndpoint({
                    applicationID,
                }) });
            cfg.data = {
                script,
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationTriggerScriptCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationTriggerScript(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationTriggerScriptEndpoint(a) {
        const { applicationID, } = a || {};
        return `/application/${applicationID}/trigger`;
    }
    // Reorder applications
    applicationReorder(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { applicationIDs, } = a || {};
            if (!applicationIDs) {
                throw Error('field applicationIDs is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.applicationReorderEndpoint() });
            cfg.data = {
                applicationIDs,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    applicationReorderCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.applicationReorder(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    applicationReorderEndpoint() {
        return '/application/reorder';
    }
    // List labels
    labelList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, name, value, limit, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.labelListEndpoint() });
            cfg.params = {
                kind,
                name,
                value,
                limit,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    labelListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.labelList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    labelListEndpoint() {
        return 'undefined/label/';
    }
    // Retrieve defined permissions
    permissionsList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    permissionsListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsListEndpoint() {
        return '/permissions/';
    }
    // Effective rules for current user
    permissionsEffective(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resource, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsEffectiveEndpoint() });
            cfg.params = {
                resource,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    permissionsEffectiveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsEffective(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsEffectiveEndpoint() {
        return '/permissions/effective';
    }
    // Evaluate rules for given user/role combo
    permissionsTrace(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resource, userID, roleID, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsTraceEndpoint() });
            cfg.params = {
                resource,
                userID,
                roleID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    permissionsTraceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsTrace(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsTraceEndpoint() {
        return '/permissions/trace';
    }
    // Retrieve role permissions
    permissionsRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, resource, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsReadEndpoint({
                    roleID,
                }) });
            cfg.params = {
                resource,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    permissionsReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsReadEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // Remove all defined role permissions
    permissionsDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.permissionsDeleteEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    permissionsDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsDeleteEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // Update permission settings
    permissionsUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, rules, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!rules) {
                throw Error('field rules is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.permissionsUpdateEndpoint({
                    roleID,
                }) });
            cfg.data = {
                rules,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    permissionsUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsUpdateEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // List/read reminders
    reminderList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reminderID, resource, assignedTo, scheduledFrom, scheduledUntil, scheduledOnly, excludeDismissed, includeDeleted, limit, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.reminderListEndpoint() });
            cfg.params = {
                reminderID,
                resource,
                assignedTo,
                scheduledFrom,
                scheduledUntil,
                scheduledOnly,
                excludeDismissed,
                includeDeleted,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reminderListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reminderList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reminderListEndpoint() {
        return '/reminder/';
    }
    // Add new reminder
    reminderCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resource, assignedTo, payload, remindAt, } = a || {};
            if (!resource) {
                throw Error('field resource is empty');
            }
            if (!assignedTo) {
                throw Error('field assignedTo is empty');
            }
            if (!payload) {
                throw Error('field payload is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.reminderCreateEndpoint() });
            cfg.data = {
                resource,
                assignedTo,
                payload,
                remindAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reminderCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reminderCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reminderCreateEndpoint() {
        return '/reminder/';
    }
    // Update reminder
    reminderUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reminderID, resource, assignedTo, payload, remindAt, } = a || {};
            if (!reminderID) {
                throw Error('field reminderID is empty');
            }
            if (!resource) {
                throw Error('field resource is empty');
            }
            if (!assignedTo) {
                throw Error('field assignedTo is empty');
            }
            if (!payload) {
                throw Error('field payload is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.reminderUpdateEndpoint({
                    reminderID,
                }) });
            cfg.data = {
                resource,
                assignedTo,
                payload,
                remindAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reminderUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reminderUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reminderUpdateEndpoint(a) {
        const { reminderID, } = a || {};
        return `/reminder/${reminderID}`;
    }
    // Read reminder by ID
    reminderRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reminderID, } = a || {};
            if (!reminderID) {
                throw Error('field reminderID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.reminderReadEndpoint({
                    reminderID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reminderReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reminderRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reminderReadEndpoint(a) {
        const { reminderID, } = a || {};
        return `/reminder/${reminderID}`;
    }
    // Delete reminder
    reminderDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reminderID, } = a || {};
            if (!reminderID) {
                throw Error('field reminderID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.reminderDeleteEndpoint({
                    reminderID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reminderDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reminderDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reminderDeleteEndpoint(a) {
        const { reminderID, } = a || {};
        return `/reminder/${reminderID}`;
    }
    // Dismiss reminder
    reminderDismiss(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reminderID, } = a || {};
            if (!reminderID) {
                throw Error('field reminderID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.reminderDismissEndpoint({
                    reminderID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reminderDismissCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reminderDismiss(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reminderDismissEndpoint(a) {
        const { reminderID, } = a || {};
        return `/reminder/${reminderID}/dismiss`;
    }
    // Undismiss reminder
    reminderUndismiss(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reminderID, } = a || {};
            if (!reminderID) {
                throw Error('field reminderID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.reminderUndismissEndpoint({
                    reminderID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reminderUndismissCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reminderUndismiss(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reminderUndismissEndpoint(a) {
        const { reminderID, } = a || {};
        return `/reminder/${reminderID}/undismiss`;
    }
    // Snooze reminder
    reminderSnooze(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reminderID, remindAt, } = a || {};
            if (!reminderID) {
                throw Error('field reminderID is empty');
            }
            if (!remindAt) {
                throw Error('field remindAt is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.reminderSnoozeEndpoint({
                    reminderID,
                }) });
            cfg.data = {
                remindAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reminderSnoozeCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reminderSnooze(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reminderSnoozeEndpoint(a) {
        const { reminderID, } = a || {};
        return `/reminder/${reminderID}/snooze`;
    }
    // List/read notifications
    notificationList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { notificationID, kind, read, deleted, limit, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.notificationListEndpoint() });
            cfg.params = {
                notificationID,
                kind,
                read,
                deleted,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    notificationListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationListEndpoint() {
        return '/notification/';
    }
    // Add new notification
    notificationCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, config, recipient, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!config) {
                throw Error('field config is empty');
            }
            if (!recipient) {
                throw Error('field recipient is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.notificationCreateEndpoint() });
            cfg.data = {
                kind,
                config,
                recipient,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    notificationCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationCreateEndpoint() {
        return '/notification/';
    }
    // Update notification
    notificationUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { notificationID, kind, config, recipient, } = a || {};
            if (!notificationID) {
                throw Error('field notificationID is empty');
            }
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!config) {
                throw Error('field config is empty');
            }
            if (!recipient) {
                throw Error('field recipient is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.notificationUpdateEndpoint({
                    notificationID,
                }) });
            cfg.data = {
                kind,
                config,
                recipient,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    notificationUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationUpdateEndpoint(a) {
        const { notificationID, } = a || {};
        return `/notification/${notificationID}`;
    }
    // Read notification by ID
    notificationRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { notificationID, } = a || {};
            if (!notificationID) {
                throw Error('field notificationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.notificationReadEndpoint({
                    notificationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    notificationReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationReadEndpoint(a) {
        const { notificationID, } = a || {};
        return `/notification/${notificationID}`;
    }
    // Delete notification
    notificationDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { notificationID, } = a || {};
            if (!notificationID) {
                throw Error('field notificationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.notificationDeleteEndpoint({
                    notificationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    notificationDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationDeleteEndpoint(a) {
        const { notificationID, } = a || {};
        return `/notification/${notificationID}`;
    }
    // Mark notification as read
    notificationMarkAsRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { notificationID, } = a || {};
            if (!notificationID) {
                throw Error('field notificationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.notificationMarkAsReadEndpoint({
                    notificationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    notificationMarkAsReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationMarkAsRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationMarkAsReadEndpoint(a) {
        const { notificationID, } = a || {};
        return `/notification/${notificationID}/read`;
    }
    // Mark notification as unread
    notificationMarkAsUnread(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { notificationID, } = a || {};
            if (!notificationID) {
                throw Error('field notificationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.notificationMarkAsUnreadEndpoint({
                    notificationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    notificationMarkAsUnreadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationMarkAsUnread(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationMarkAsUnreadEndpoint(a) {
        const { notificationID, } = a || {};
        return `/notification/${notificationID}/unread`;
    }
    // Mark all notifications as read for current user
    notificationMarkAllAsRead() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.notificationMarkAllAsReadEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    notificationMarkAllAsReadCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationMarkAllAsRead(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationMarkAllAsReadEndpoint() {
        return '/notification/all/read';
    }
    // Mark all notifications as unread for current user
    notificationMarkAllAsUnread() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.notificationMarkAllAsUnreadEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    notificationMarkAllAsUnreadCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationMarkAllAsUnread(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationMarkAllAsUnreadEndpoint() {
        return '/notification/all/unread';
    }
    // Attachment details
    attachmentRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, attachmentID, sign, userID, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!attachmentID) {
                throw Error('field attachmentID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.attachmentReadEndpoint({
                    kind, attachmentID,
                }) });
            cfg.params = {
                sign,
                userID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    attachmentReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.attachmentRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    attachmentReadEndpoint(a) {
        const { kind, attachmentID, } = a || {};
        return `/attachment/${kind}/${attachmentID}`;
    }
    // Delete attachment
    attachmentDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, attachmentID, sign, userID, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!attachmentID) {
                throw Error('field attachmentID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.attachmentDeleteEndpoint({
                    kind, attachmentID,
                }) });
            cfg.params = {
                sign,
                userID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    attachmentDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.attachmentDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    attachmentDeleteEndpoint(a) {
        const { kind, attachmentID, } = a || {};
        return `/attachment/${kind}/${attachmentID}`;
    }
    // Serves attached file
    attachmentOriginal(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, attachmentID, name, sign, userID, download, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!attachmentID) {
                throw Error('field attachmentID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.attachmentOriginalEndpoint({
                    kind, attachmentID, name,
                }) });
            cfg.params = {
                sign,
                userID,
                download,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    attachmentOriginalCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.attachmentOriginal(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    attachmentOriginalEndpoint(a) {
        const { kind, attachmentID, name, } = a || {};
        return `/attachment/${kind}/${attachmentID}/original/${name}`;
    }
    // Serves preview of an attached file
    attachmentPreview(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, attachmentID, ext, sign, userID, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!attachmentID) {
                throw Error('field attachmentID is empty');
            }
            if (!ext) {
                throw Error('field ext is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.attachmentPreviewEndpoint({
                    kind, attachmentID, ext,
                }) });
            cfg.params = {
                sign,
                userID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    attachmentPreviewCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.attachmentPreview(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    attachmentPreviewEndpoint(a) {
        const { kind, attachmentID, ext, } = a || {};
        return `/attachment/${kind}/${attachmentID}/preview.${ext}`;
    }
    // List templates
    templateList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { query, handle, type, ownerID, partial, deleted, labels, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.templateListEndpoint() });
            cfg.params = {
                query,
                handle,
                type,
                ownerID,
                partial,
                deleted,
                labels,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    templateListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.templateList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    templateListEndpoint() {
        return '/template/';
    }
    // Create template
    templateCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { handle, language, type, partial, meta, template, ownerID, labels, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.templateCreateEndpoint() });
            cfg.data = {
                handle,
                language,
                type,
                partial,
                meta,
                template,
                ownerID,
                labels,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    templateCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.templateCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    templateCreateEndpoint() {
        return '/template/';
    }
    // Read template
    templateRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { templateID, } = a || {};
            if (!templateID) {
                throw Error('field templateID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.templateReadEndpoint({
                    templateID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    templateReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.templateRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    templateReadEndpoint(a) {
        const { templateID, } = a || {};
        return `/template/${templateID}`;
    }
    // Update template
    templateUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { templateID, handle, language, type, partial, meta, template, ownerID, labels, updatedAt, } = a || {};
            if (!templateID) {
                throw Error('field templateID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.templateUpdateEndpoint({
                    templateID,
                }) });
            cfg.data = {
                handle,
                language,
                type,
                partial,
                meta,
                template,
                ownerID,
                labels,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    templateUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.templateUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    templateUpdateEndpoint(a) {
        const { templateID, } = a || {};
        return `/template/${templateID}`;
    }
    // Delete template
    templateDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { templateID, } = a || {};
            if (!templateID) {
                throw Error('field templateID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.templateDeleteEndpoint({
                    templateID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    templateDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.templateDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    templateDeleteEndpoint(a) {
        const { templateID, } = a || {};
        return `/template/${templateID}`;
    }
    // Undelete template
    templateUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { templateID, } = a || {};
            if (!templateID) {
                throw Error('field templateID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.templateUndeleteEndpoint({
                    templateID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    templateUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.templateUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    templateUndeleteEndpoint(a) {
        const { templateID, } = a || {};
        return `/template/${templateID}/undelete`;
    }
    // Render drivers
    templateRenderDrivers() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.templateRenderDriversEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    templateRenderDriversCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.templateRenderDrivers(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    templateRenderDriversEndpoint() {
        return '/template/render/drivers';
    }
    // Render template
    templateRender(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { templateID, filename, ext, variables, options, } = a || {};
            if (!templateID) {
                throw Error('field templateID is empty');
            }
            if (!filename) {
                throw Error('field filename is empty');
            }
            if (!ext) {
                throw Error('field ext is empty');
            }
            if (!variables) {
                throw Error('field variables is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.templateRenderEndpoint({
                    templateID, filename, ext,
                }) });
            cfg.data = {
                variables,
                options,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    templateRenderCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.templateRender(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    templateRenderEndpoint(a) {
        const { templateID, filename, ext, } = a || {};
        return `/template/${templateID}/render/${filename}.${ext}`;
    }
    // List reports
    reportList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { handle, query, deleted, labels, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.reportListEndpoint() });
            cfg.params = {
                handle,
                query,
                deleted,
                labels,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reportListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reportList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reportListEndpoint() {
        return '/reports/';
    }
    // Create report
    reportCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { handle, meta, scenarios, sources, blocks, labels, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.reportCreateEndpoint() });
            cfg.data = {
                handle,
                meta,
                scenarios,
                sources,
                blocks,
                labels,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reportCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reportCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reportCreateEndpoint() {
        return '/reports/';
    }
    // Update report
    reportUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reportID, handle, meta, scenarios, sources, blocks, labels, updatedAt, } = a || {};
            if (!reportID) {
                throw Error('field reportID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.reportUpdateEndpoint({
                    reportID,
                }) });
            cfg.data = {
                handle,
                meta,
                scenarios,
                sources,
                blocks,
                labels,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reportUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reportUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reportUpdateEndpoint(a) {
        const { reportID, } = a || {};
        return `/reports/${reportID}`;
    }
    // Read report details
    reportRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reportID, } = a || {};
            if (!reportID) {
                throw Error('field reportID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.reportReadEndpoint({
                    reportID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reportReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reportRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reportReadEndpoint(a) {
        const { reportID, } = a || {};
        return `/reports/${reportID}`;
    }
    // Remove report
    reportDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reportID, } = a || {};
            if (!reportID) {
                throw Error('field reportID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.reportDeleteEndpoint({
                    reportID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reportDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reportDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reportDeleteEndpoint(a) {
        const { reportID, } = a || {};
        return `/reports/${reportID}`;
    }
    // Undelete report
    reportUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reportID, } = a || {};
            if (!reportID) {
                throw Error('field reportID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.reportUndeleteEndpoint({
                    reportID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reportUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reportUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reportUndeleteEndpoint(a) {
        const { reportID, } = a || {};
        return `/reports/${reportID}/undelete`;
    }
    // Describe report
    reportDescribe(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sources, steps, describe, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.reportDescribeEndpoint() });
            cfg.data = {
                sources,
                steps,
                describe,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reportDescribeCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reportDescribe(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reportDescribeEndpoint() {
        return '/reports/describe';
    }
    // Run report
    reportRun(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { reportID, frames, } = a || {};
            if (!reportID) {
                throw Error('field reportID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.reportRunEndpoint({
                    reportID,
                }) });
            cfg.data = {
                frames,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    reportRunCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.reportRun(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    reportRunEndpoint(a) {
        const { reportID, } = a || {};
        return `/reports/${reportID}/run`;
    }
    // List system statistics
    statsList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.statsListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    statsListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.statsList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    statsListEndpoint() {
        return '/stats/';
    }
    // List all available automation scripts for system resources
    automationList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resourceTypePrefixes, resourceTypes, eventTypes, excludeInvalid, excludeClientScripts, excludeServerScripts, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.automationListEndpoint() });
            cfg.params = {
                resourceTypePrefixes,
                resourceTypes,
                eventTypes,
                excludeInvalid,
                excludeClientScripts,
                excludeServerScripts,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    automationListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.automationList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    automationListEndpoint() {
        return '/automation/';
    }
    // Serves client scripts bundle
    automationBundle(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { bundle, type, ext, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.automationBundleEndpoint({
                    bundle, type, ext,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    automationBundleCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.automationBundle(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    automationBundleEndpoint(a) {
        const { bundle, type, ext, } = a || {};
        return `/automation/${bundle}-${type}.${ext}`;
    }
    // Triggers execution of a specific script on a system service level
    automationTriggerScript(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { script, args, } = a || {};
            if (!script) {
                throw Error('field script is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.automationTriggerScriptEndpoint() });
            cfg.data = {
                script,
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    automationTriggerScriptCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.automationTriggerScript(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    automationTriggerScriptEndpoint() {
        return '/automation/trigger';
    }
    // Action log events
    actionlogList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { from, to, beforeActionID, resource, action, actorID, limit, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.actionlogListEndpoint() });
            cfg.params = {
                from,
                to,
                beforeActionID,
                resource,
                action,
                actorID,
                limit,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    actionlogListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.actionlogList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    actionlogListEndpoint() {
        return '/actionlog/';
    }
    // Messaging queues
    queuesList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { query, limit, incTotal, pageCursor, sort, deleted, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.queuesListEndpoint() });
            cfg.params = {
                query,
                limit,
                incTotal,
                pageCursor,
                sort,
                deleted,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    queuesListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.queuesList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    queuesListEndpoint() {
        return '/queues/';
    }
    // Create messaging queue
    queuesCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { queue, consumer, meta, } = a || {};
            if (!queue) {
                throw Error('field queue is empty');
            }
            if (!consumer) {
                throw Error('field consumer is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.queuesCreateEndpoint() });
            cfg.data = {
                queue,
                consumer,
                meta,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    queuesCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.queuesCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    queuesCreateEndpoint() {
        return '/queues';
    }
    // Messaging queue details
    queuesRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { queueID, } = a || {};
            if (!queueID) {
                throw Error('field queueID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.queuesReadEndpoint({
                    queueID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    queuesReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.queuesRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    queuesReadEndpoint(a) {
        const { queueID, } = a || {};
        return `/queues/${queueID}`;
    }
    // Update queue details
    queuesUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { queueID, queue, consumer, meta, updatedAt, } = a || {};
            if (!queueID) {
                throw Error('field queueID is empty');
            }
            if (!queue) {
                throw Error('field queue is empty');
            }
            if (!consumer) {
                throw Error('field consumer is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.queuesUpdateEndpoint({
                    queueID,
                }) });
            cfg.data = {
                queue,
                consumer,
                meta,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    queuesUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.queuesUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    queuesUpdateEndpoint(a) {
        const { queueID, } = a || {};
        return `/queues/${queueID}`;
    }
    // Messaging queue delete
    queuesDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { queueID, } = a || {};
            if (!queueID) {
                throw Error('field queueID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.queuesDeleteEndpoint({
                    queueID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    queuesDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.queuesDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    queuesDeleteEndpoint(a) {
        const { queueID, } = a || {};
        return `/queues/${queueID}`;
    }
    // Messaging queue undelete
    queuesUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { queueID, } = a || {};
            if (!queueID) {
                throw Error('field queueID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.queuesUndeleteEndpoint({
                    queueID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    queuesUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.queuesUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    queuesUndeleteEndpoint(a) {
        const { queueID, } = a || {};
        return `/queues/${queueID}/undelete`;
    }
    // List routes
    apigwRouteList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { routeID, query, deleted, disabled, labels, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.apigwRouteListEndpoint() });
            cfg.params = {
                routeID,
                query,
                deleted,
                disabled,
                labels,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwRouteListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwRouteList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwRouteListEndpoint() {
        return '/apigw/route/';
    }
    // Create route
    apigwRouteCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { endpoint, method, enabled, group, meta, } = a || {};
            if (!endpoint) {
                throw Error('field endpoint is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.apigwRouteCreateEndpoint() });
            cfg.data = {
                endpoint,
                method,
                enabled,
                group,
                meta,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwRouteCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwRouteCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwRouteCreateEndpoint() {
        return '/apigw/route';
    }
    // Update route details
    apigwRouteUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { routeID, endpoint, method, enabled, group, meta, updatedAt, } = a || {};
            if (!routeID) {
                throw Error('field routeID is empty');
            }
            if (!endpoint) {
                throw Error('field endpoint is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.apigwRouteUpdateEndpoint({
                    routeID,
                }) });
            cfg.data = {
                endpoint,
                method,
                enabled,
                group,
                meta,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwRouteUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwRouteUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwRouteUpdateEndpoint(a) {
        const { routeID, } = a || {};
        return `/apigw/route/${routeID}`;
    }
    // Read route details
    apigwRouteRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { routeID, } = a || {};
            if (!routeID) {
                throw Error('field routeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.apigwRouteReadEndpoint({
                    routeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwRouteReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwRouteRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwRouteReadEndpoint(a) {
        const { routeID, } = a || {};
        return `/apigw/route/${routeID}`;
    }
    // Remove route
    apigwRouteDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { routeID, } = a || {};
            if (!routeID) {
                throw Error('field routeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.apigwRouteDeleteEndpoint({
                    routeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwRouteDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwRouteDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwRouteDeleteEndpoint(a) {
        const { routeID, } = a || {};
        return `/apigw/route/${routeID}`;
    }
    // Undelete route
    apigwRouteUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { routeID, } = a || {};
            if (!routeID) {
                throw Error('field routeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.apigwRouteUndeleteEndpoint({
                    routeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwRouteUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwRouteUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwRouteUndeleteEndpoint(a) {
        const { routeID, } = a || {};
        return `/apigw/route/${routeID}/undelete`;
    }
    // List filters
    apigwFilterList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { routeID, deleted, disabled, limit, pageCursor, sort, } = a || {};
            if (!routeID) {
                throw Error('field routeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.apigwFilterListEndpoint() });
            cfg.params = {
                routeID,
                deleted,
                disabled,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwFilterListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwFilterList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwFilterListEndpoint() {
        return '/apigw/filter/';
    }
    // Create filter
    apigwFilterCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { routeID, weight, kind, ref, enabled, params, } = a || {};
            if (!routeID) {
                throw Error('field routeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.apigwFilterCreateEndpoint() });
            cfg.data = {
                routeID,
                weight,
                kind,
                ref,
                enabled,
                params,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwFilterCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwFilterCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwFilterCreateEndpoint() {
        return '/apigw/filter';
    }
    // Update filter details
    apigwFilterUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { filterID, routeID, weight, kind, ref, enabled, params, updatedAt, } = a || {};
            if (!filterID) {
                throw Error('field filterID is empty');
            }
            if (!routeID) {
                throw Error('field routeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.apigwFilterUpdateEndpoint({
                    filterID,
                }) });
            cfg.data = {
                routeID,
                weight,
                kind,
                ref,
                enabled,
                params,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwFilterUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwFilterUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwFilterUpdateEndpoint(a) {
        const { filterID, } = a || {};
        return `/apigw/filter/${filterID}`;
    }
    // Read filter details
    apigwFilterRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { filterID, } = a || {};
            if (!filterID) {
                throw Error('field filterID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.apigwFilterReadEndpoint({
                    filterID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwFilterReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwFilterRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwFilterReadEndpoint(a) {
        const { filterID, } = a || {};
        return `/apigw/filter/${filterID}`;
    }
    // Remove filter
    apigwFilterDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { filterID, } = a || {};
            if (!filterID) {
                throw Error('field filterID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.apigwFilterDeleteEndpoint({
                    filterID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwFilterDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwFilterDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwFilterDeleteEndpoint(a) {
        const { filterID, } = a || {};
        return `/apigw/filter/${filterID}`;
    }
    // Undelete filter
    apigwFilterUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { filterID, } = a || {};
            if (!filterID) {
                throw Error('field filterID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.apigwFilterUndeleteEndpoint({
                    filterID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwFilterUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwFilterUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwFilterUndeleteEndpoint(a) {
        const { filterID, } = a || {};
        return `/apigw/filter/${filterID}/undelete`;
    }
    // Filter definitions
    apigwFilterDefFilter(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.apigwFilterDefFilterEndpoint() });
            cfg.params = {
                kind,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwFilterDefFilterCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwFilterDefFilter(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwFilterDefFilterEndpoint() {
        return '/apigw/filter/def';
    }
    // Proxy auth definitions
    apigwFilterDefProxyAuth() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.apigwFilterDefProxyAuthEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwFilterDefProxyAuthCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwFilterDefProxyAuth(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwFilterDefProxyAuthEndpoint() {
        return '/apigw/filter/proxy_auth/def';
    }
    // List aggregated list of routes
    apigwProfilerAggregation(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { path, before, sort, limit, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.apigwProfilerAggregationEndpoint() });
            cfg.params = {
                path,
                before,
                sort,
                limit,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwProfilerAggregationCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwProfilerAggregation(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwProfilerAggregationEndpoint() {
        return '/apigw/profiler/';
    }
    // List hits per route
    apigwProfilerRoute(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { routeID, path, before, sort, limit, } = a || {};
            if (!routeID) {
                throw Error('field routeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.apigwProfilerRouteEndpoint({
                    routeID,
                }) });
            cfg.params = {
                path,
                before,
                sort,
                limit,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwProfilerRouteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwProfilerRoute(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwProfilerRouteEndpoint(a) {
        const { routeID, } = a || {};
        return `/apigw/profiler/route/${routeID}`;
    }
    // Hit details
    apigwProfilerHit(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { hitID, } = a || {};
            if (!hitID) {
                throw Error('field hitID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.apigwProfilerHitEndpoint({
                    hitID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwProfilerHitCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwProfilerHit(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwProfilerHitEndpoint(a) {
        const { hitID, } = a || {};
        return `/apigw/profiler/hit/${hitID}`;
    }
    // Purge all profiler hits
    apigwProfilerPurgeAll() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.apigwProfilerPurgeAllEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwProfilerPurgeAllCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwProfilerPurgeAll(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwProfilerPurgeAllEndpoint() {
        return '/apigw/profiler/purge';
    }
    // Purge route profiler hits
    apigwProfilerPurge(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { routeID, } = a || {};
            if (!routeID) {
                throw Error('field routeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.apigwProfilerPurgeEndpoint({
                    routeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    apigwProfilerPurgeCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.apigwProfilerPurge(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    apigwProfilerPurgeEndpoint(a) {
        const { routeID, } = a || {};
        return `/apigw/profiler/purge/${routeID}`;
    }
    // List resources translations
    localeListResource(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { lang, resource, resourceType, ownerID, deleted, limit, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.localeListResourceEndpoint() });
            cfg.params = {
                lang,
                resource,
                resourceType,
                ownerID,
                deleted,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    localeListResourceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.localeListResource(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    localeListResourceEndpoint() {
        return '/locale/resource';
    }
    // Create resource translation
    localeCreateResource(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { lang, resource, key, place, message, ownerID, } = a || {};
            if (!lang) {
                throw Error('field lang is empty');
            }
            if (!resource) {
                throw Error('field resource is empty');
            }
            if (!key) {
                throw Error('field key is empty');
            }
            if (!message) {
                throw Error('field message is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.localeCreateResourceEndpoint() });
            cfg.data = {
                lang,
                resource,
                key,
                place,
                message,
                ownerID,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    localeCreateResourceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.localeCreateResource(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    localeCreateResourceEndpoint() {
        return '/locale/resource';
    }
    // Update resource translation
    localeUpdateResource(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { translationID, lang, resource, key, place, message, ownerID, updatedAt, } = a || {};
            if (!translationID) {
                throw Error('field translationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.localeUpdateResourceEndpoint({
                    translationID,
                }) });
            cfg.data = {
                lang,
                resource,
                key,
                place,
                message,
                ownerID,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    localeUpdateResourceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.localeUpdateResource(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    localeUpdateResourceEndpoint(a) {
        const { translationID, } = a || {};
        return `/locale/resource/${translationID}`;
    }
    // Read resource translation details
    localeReadResource(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { translationID, } = a || {};
            if (!translationID) {
                throw Error('field translationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.localeReadResourceEndpoint({
                    translationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    localeReadResourceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.localeReadResource(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    localeReadResourceEndpoint(a) {
        const { translationID, } = a || {};
        return `/locale/resource/${translationID}`;
    }
    // Remove resource translation
    localeDeleteResource(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { translationID, } = a || {};
            if (!translationID) {
                throw Error('field translationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.localeDeleteResourceEndpoint({
                    translationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    localeDeleteResourceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.localeDeleteResource(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    localeDeleteResourceEndpoint(a) {
        const { translationID, } = a || {};
        return `/locale/resource/${translationID}`;
    }
    // Undelete resource translation
    localeUndeleteResource(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { translationID, } = a || {};
            if (!translationID) {
                throw Error('field translationID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.localeUndeleteResourceEndpoint({
                    translationID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    localeUndeleteResourceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.localeUndeleteResource(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    localeUndeleteResourceEndpoint(a) {
        const { translationID, } = a || {};
        return `/locale/resource/${translationID}/undelete`;
    }
    // List all available languages
    localeList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.localeListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    localeListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.localeList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    localeListEndpoint() {
        return '/locale/';
    }
    // List all available translation in a language for a specific webapp
    localeGet(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { lang, application, } = a || {};
            if (!lang) {
                throw Error('field lang is empty');
            }
            if (!application) {
                throw Error('field application is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.localeGetEndpoint({
                    lang, application,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    localeGetCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.localeGet(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    localeGetEndpoint(a) {
        const { lang, application, } = a || {};
        return `/locale/${lang}/${application}`;
    }
    // List connections for data privacy
    dataPrivacyConnectionList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { connectionID, handle, type, deleted, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dataPrivacyConnectionListEndpoint() });
            cfg.params = {
                connectionID,
                handle,
                type,
                deleted,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dataPrivacyConnectionListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dataPrivacyConnectionList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dataPrivacyConnectionListEndpoint() {
        return '/data-privacy/connection/';
    }
    // List data privacy requests
    dataPrivacyRequestList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { requestedBy, query, kind, status, limit, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dataPrivacyRequestListEndpoint() });
            cfg.params = {
                requestedBy,
                query,
                kind,
                status,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dataPrivacyRequestListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dataPrivacyRequestList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dataPrivacyRequestListEndpoint() {
        return '/data-privacy/requests/';
    }
    // Create data privacy request
    dataPrivacyRequestCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, payload, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.dataPrivacyRequestCreateEndpoint() });
            cfg.data = {
                kind,
                payload,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dataPrivacyRequestCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dataPrivacyRequestCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dataPrivacyRequestCreateEndpoint() {
        return '/data-privacy/requests/';
    }
    // Get details about specific request
    dataPrivacyRequestRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { requestID, } = a || {};
            if (!requestID) {
                throw Error('field requestID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dataPrivacyRequestReadEndpoint({
                    requestID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dataPrivacyRequestReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dataPrivacyRequestRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dataPrivacyRequestReadEndpoint(a) {
        const { requestID, } = a || {};
        return `/data-privacy/requests/${requestID}`;
    }
    // Update data privacy request status
    dataPrivacyRequestUpdateStatus(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { requestID, status, } = a || {};
            if (!requestID) {
                throw Error('field requestID is empty');
            }
            if (!status) {
                throw Error('field status is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.dataPrivacyRequestUpdateStatusEndpoint({
                    requestID, status,
                }) });
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dataPrivacyRequestUpdateStatusCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dataPrivacyRequestUpdateStatus(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dataPrivacyRequestUpdateStatusEndpoint(a) {
        const { requestID, status, } = a || {};
        return `/data-privacy/requests/${requestID}/status/${status}`;
    }
    // List data privacy request comments
    dataPrivacyRequestCommentList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { requestID, limit, pageCursor, sort, } = a || {};
            if (!requestID) {
                throw Error('field requestID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dataPrivacyRequestCommentListEndpoint({
                    requestID,
                }) });
            cfg.params = {
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dataPrivacyRequestCommentListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dataPrivacyRequestCommentList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dataPrivacyRequestCommentListEndpoint(a) {
        const { requestID, } = a || {};
        return `/data-privacy/requests/${requestID}/comments/`;
    }
    // Create data privacy request comment
    dataPrivacyRequestCommentCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { requestID, comment, } = a || {};
            if (!requestID) {
                throw Error('field requestID is empty');
            }
            if (!comment) {
                throw Error('field comment is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.dataPrivacyRequestCommentCreateEndpoint({
                    requestID,
                }) });
            cfg.data = {
                comment,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    dataPrivacyRequestCommentCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dataPrivacyRequestCommentCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dataPrivacyRequestCommentCreateEndpoint(a) {
        const { requestID, } = a || {};
        return `/data-privacy/requests/${requestID}/comments/`;
    }
    // Check SMTP server configuration settings
    smtpConfigurationCheckerCheck(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { host, port, recipients, username, password, tlsInsecure, tlsServerName, } = a || {};
            if (!host) {
                throw Error('field host is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.smtpConfigurationCheckerCheckEndpoint() });
            cfg.data = {
                host,
                port,
                recipients,
                username,
                password,
                tlsInsecure,
                tlsServerName,
            };
            return this.api().request(cfg).then(result => stdResolve$4(result));
        });
    }
    smtpConfigurationCheckerCheckCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.smtpConfigurationCheckerCheck(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    smtpConfigurationCheckerCheckEndpoint() {
        return '/smtp/configuration-checker/';
    }
}

/* eslint-disable padded-blocks */
function stdResolve$3(response) {
    if (response.data.error) {
        return Promise.reject(response.data.error);
    }
    else {
        return response.data.response;
    }
}
class Compose {
    constructor({ baseURL, headers, accessTokenFn }) {
        this.headers = {};
        this.baseURL = baseURL;
        this.accessTokenFn = accessTokenFn;
        this.headers = {
            /**
             * All we send is JSON
             */
            'Content-Type': 'application/json',
        };
        this.setHeaders(headers);
    }
    setAccessTokenFn(fn) {
        this.accessTokenFn = fn;
        return this;
    }
    setHeaders(headers) {
        if (typeof headers === 'object') {
            this.headers = headers;
        }
        return this;
    }
    setHeader(name, value) {
        if (value === undefined) {
            delete this.headers[name];
        }
        else {
            this.headers[name] = value;
        }
        return this;
    }
    api() {
        const headers = Object.assign({}, this.headers);
        const accessToken = this.accessTokenFn ? this.accessTokenFn() : undefined;
        if (accessToken) {
            headers.Authorization = 'Bearer ' + accessToken;
        }
        return axios.create({
            withCredentials: true,
            baseURL: this.baseURL,
            headers,
        });
    }
    // List namespaces
    namespaceList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { query, slug, limit, incTotal, labels, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.namespaceListEndpoint() });
            cfg.params = {
                query,
                slug,
                limit,
                incTotal,
                labels,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceListEndpoint() {
        return '/namespace/';
    }
    // Create namespace
    namespaceCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { name, labels, slug, enabled, meta, } = a || {};
            if (!name) {
                throw Error('field name is empty');
            }
            if (!meta) {
                throw Error('field meta is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.namespaceCreateEndpoint() });
            cfg.data = {
                name,
                labels,
                slug,
                enabled,
                meta,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceCreateEndpoint() {
        return '/namespace/';
    }
    // Read namespace
    namespaceRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.namespaceReadEndpoint({
                    namespaceID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceReadEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}`;
    }
    // Update namespace
    namespaceUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, name, slug, enabled, meta, labels, updatedAt, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            if (!meta) {
                throw Error('field meta is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.namespaceUpdateEndpoint({
                    namespaceID,
                }) });
            cfg.data = {
                name,
                slug,
                enabled,
                meta,
                labels,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceUpdateEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}`;
    }
    // Delete namespace
    namespaceDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.namespaceDeleteEndpoint({
                    namespaceID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceDeleteEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}`;
    }
    // Upload namespace assets
    namespaceUpload(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { upload, } = a || {};
            if (!upload) {
                throw Error('field upload is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.namespaceUploadEndpoint() });
            cfg.data = {
                upload,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceUploadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceUpload(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceUploadEndpoint() {
        return '/namespace/upload';
    }
    // Clone compose namespace
    namespaceClone(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, name, slug, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.namespaceCloneEndpoint({
                    namespaceID,
                }) });
            cfg.data = {
                name,
                slug,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceCloneCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceClone(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceCloneEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/clone`;
    }
    // Export compose namespace
    namespaceExport(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, filename, ext, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!filename) {
                throw Error('field filename is empty');
            }
            if (!ext) {
                throw Error('field ext is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.namespaceExportEndpoint({
                    namespaceID, filename, ext,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceExportCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceExport(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceExportEndpoint(a) {
        const { namespaceID, filename, ext, } = a || {};
        return `/namespace/${namespaceID}/export/${filename}.zip`;
    }
    // Initiate namespace import session
    namespaceImportInit(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { upload, } = a || {};
            if (!upload) {
                throw Error('field upload is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.namespaceImportInitEndpoint() });
            cfg.data = {
                upload,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceImportInitCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceImportInit(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceImportInitEndpoint() {
        return '/namespace/import';
    }
    // Run namespace import
    namespaceImportRun(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sessionID, name, slug, } = a || {};
            if (!sessionID) {
                throw Error('field sessionID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.namespaceImportRunEndpoint({
                    sessionID,
                }) });
            cfg.data = {
                name,
                slug,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceImportRunCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceImportRun(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceImportRunEndpoint(a) {
        const { sessionID, } = a || {};
        return `/namespace/import/${sessionID}`;
    }
    // Fire compose:namespace trigger
    namespaceTriggerScript(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, script, args, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!script) {
                throw Error('field script is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.namespaceTriggerScriptEndpoint({
                    namespaceID,
                }) });
            cfg.data = {
                script,
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceTriggerScriptCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceTriggerScript(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceTriggerScriptEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/trigger`;
    }
    // List translation
    namespaceListTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.namespaceListTranslationsEndpoint({
                    namespaceID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceListTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceListTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceListTranslationsEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/translation`;
    }
    // Update translation
    namespaceUpdateTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, translations, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!translations) {
                throw Error('field translations is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.namespaceUpdateTranslationsEndpoint({
                    namespaceID,
                }) });
            cfg.data = {
                translations,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    namespaceUpdateTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.namespaceUpdateTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    namespaceUpdateTranslationsEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/translation`;
    }
    // List available pages
    pageList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, selfID, moduleID, query, handle, labels, limit, pageCursor, sort, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.pageListEndpoint({
                    namespaceID,
                }) });
            cfg.params = {
                selfID,
                moduleID,
                query,
                handle,
                labels,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageListEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/page/`;
    }
    // Create page
    pageCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, selfID, moduleID, title, handle, description, weight, labels, visible, blocks, config, meta, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!title) {
                throw Error('field title is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.pageCreateEndpoint({
                    namespaceID,
                }) });
            cfg.data = {
                selfID,
                moduleID,
                title,
                handle,
                description,
                weight,
                labels,
                visible,
                blocks,
                config,
                meta,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageCreateEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/page/`;
    }
    // Get page details
    pageRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.pageReadEndpoint({
                    namespaceID, pageID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageReadEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}`;
    }
    // Get page all (non-record) pages, hierarchically
    pageTree(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.pageTreeEndpoint({
                    namespaceID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageTreeCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageTree(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageTreeEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/page/tree`;
    }
    // Update page
    pageUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, selfID, moduleID, title, handle, description, weight, labels, visible, blocks, config, meta, updatedAt, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!title) {
                throw Error('field title is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.pageUpdateEndpoint({
                    namespaceID, pageID,
                }) });
            cfg.data = {
                selfID,
                moduleID,
                title,
                handle,
                description,
                weight,
                labels,
                visible,
                blocks,
                config,
                meta,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageUpdateEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}`;
    }
    // Reorder pages
    pageReorder(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, selfID, pageIDs, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!selfID) {
                throw Error('field selfID is empty');
            }
            if (!pageIDs) {
                throw Error('field pageIDs is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.pageReorderEndpoint({
                    namespaceID, selfID,
                }) });
            cfg.data = {
                pageIDs,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageReorderCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageReorder(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageReorderEndpoint(a) {
        const { namespaceID, selfID, } = a || {};
        return `/namespace/${namespaceID}/page/${selfID}/reorder`;
    }
    // Delete page
    pageDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, strategy, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.pageDeleteEndpoint({
                    namespaceID, pageID,
                }) });
            cfg.params = {
                strategy,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageDeleteEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}`;
    }
    // Uploads attachment to page
    pageUpload(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, upload, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!upload) {
                throw Error('field upload is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.pageUploadEndpoint({
                    namespaceID, pageID,
                }) });
            cfg.data = {
                upload,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageUploadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageUpload(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageUploadEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/attachment`;
    }
    // Fire compose:page trigger
    pageTriggerScript(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, script, args, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!script) {
                throw Error('field script is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.pageTriggerScriptEndpoint({
                    namespaceID, pageID,
                }) });
            cfg.data = {
                script,
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageTriggerScriptCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageTriggerScript(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageTriggerScriptEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/trigger`;
    }
    // List page translation
    pageListTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.pageListTranslationsEndpoint({
                    namespaceID, pageID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageListTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageListTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageListTranslationsEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/translation`;
    }
    // Update page translation
    pageUpdateTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, translations, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!translations) {
                throw Error('field translations is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.pageUpdateTranslationsEndpoint({
                    namespaceID, pageID,
                }) });
            cfg.data = {
                translations,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageUpdateTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageUpdateTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageUpdateTranslationsEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/translation`;
    }
    // Update icon for page
    pageUpdateIcon(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, type, source, style, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!type) {
                throw Error('field type is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.pageUpdateIconEndpoint({
                    namespaceID, pageID,
                }) });
            cfg.data = {
                type,
                source,
                style,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageUpdateIconCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageUpdateIcon(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageUpdateIconEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/icon`;
    }
    // List icons
    iconList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.iconListEndpoint() });
            cfg.params = {
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    iconListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.iconList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    iconListEndpoint() {
        return '/icon/';
    }
    // Upload icon
    iconUpload(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { icon, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.iconUploadEndpoint() });
            cfg.data = {
                icon,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    iconUploadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.iconUpload(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    iconUploadEndpoint() {
        return '/icon/';
    }
    // Delete icon
    iconDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { iconID, } = a || {};
            if (!iconID) {
                throw Error('field iconID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.iconDeleteEndpoint({
                    iconID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    iconDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.iconDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    iconDeleteEndpoint(a) {
        const { iconID, } = a || {};
        return `/icon/${iconID}`;
    }
    // List available page layouts
    pageLayoutListNamespace(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, moduleID, parentID, query, handle, labels, limit, pageCursor, sort, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.pageLayoutListNamespaceEndpoint({
                    namespaceID,
                }) });
            cfg.params = {
                pageID,
                moduleID,
                parentID,
                query,
                handle,
                labels,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutListNamespaceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutListNamespace(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutListNamespaceEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/page-layout`;
    }
    // List available page layouts
    pageLayoutList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, moduleID, parentID, query, handle, labels, limit, pageCursor, sort, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.pageLayoutListEndpoint({
                    namespaceID, pageID,
                }) });
            cfg.params = {
                moduleID,
                parentID,
                query,
                handle,
                labels,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutListEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/layout/`;
    }
    // Create page layout
    pageLayoutCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, parentID, weight, moduleID, handle, meta, config, blocks, labels, ownedBy, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.pageLayoutCreateEndpoint({
                    namespaceID, pageID,
                }) });
            cfg.data = {
                parentID,
                weight,
                moduleID,
                handle,
                meta,
                config,
                blocks,
                labels,
                ownedBy,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutCreateEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/layout/`;
    }
    // Get page details
    pageLayoutRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, pageLayoutID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!pageLayoutID) {
                throw Error('field pageLayoutID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.pageLayoutReadEndpoint({
                    namespaceID, pageID, pageLayoutID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutReadEndpoint(a) {
        const { namespaceID, pageID, pageLayoutID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/layout/${pageLayoutID}`;
    }
    // Update page
    pageLayoutUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, pageLayoutID, parentID, weight, moduleID, handle, meta, config, blocks, labels, ownedBy, updatedAt, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!pageLayoutID) {
                throw Error('field pageLayoutID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.pageLayoutUpdateEndpoint({
                    namespaceID, pageID, pageLayoutID,
                }) });
            cfg.data = {
                parentID,
                weight,
                moduleID,
                handle,
                meta,
                config,
                blocks,
                labels,
                ownedBy,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutUpdateEndpoint(a) {
        const { namespaceID, pageID, pageLayoutID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/layout/${pageLayoutID}`;
    }
    // Reorder page layouts
    pageLayoutReorder(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, pageIDs, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!pageIDs) {
                throw Error('field pageIDs is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.pageLayoutReorderEndpoint({
                    namespaceID, pageID,
                }) });
            cfg.data = {
                pageIDs,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutReorderCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutReorder(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutReorderEndpoint(a) {
        const { namespaceID, pageID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/layout/reorder`;
    }
    // Delete page layout
    pageLayoutDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, pageLayoutID, strategy, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!pageLayoutID) {
                throw Error('field pageLayoutID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.pageLayoutDeleteEndpoint({
                    namespaceID, pageID, pageLayoutID,
                }) });
            cfg.params = {
                strategy,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutDeleteEndpoint(a) {
        const { namespaceID, pageID, pageLayoutID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/layout/${pageLayoutID}`;
    }
    // Undelete soft deleted Delete page layout
    pageLayoutUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, pageLayoutID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!pageLayoutID) {
                throw Error('field pageLayoutID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.pageLayoutUndeleteEndpoint({
                    namespaceID, pageID, pageLayoutID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutUndeleteEndpoint(a) {
        const { namespaceID, pageID, pageLayoutID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/layout/${pageLayoutID}/undelete`;
    }
    // List page layout translation
    pageLayoutListTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, pageLayoutID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!pageLayoutID) {
                throw Error('field pageLayoutID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.pageLayoutListTranslationsEndpoint({
                    namespaceID, pageID, pageLayoutID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutListTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutListTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutListTranslationsEndpoint(a) {
        const { namespaceID, pageID, pageLayoutID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/layout/${pageLayoutID}/translation`;
    }
    // Update page layout translation
    pageLayoutUpdateTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, pageID, pageLayoutID, translations, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!pageID) {
                throw Error('field pageID is empty');
            }
            if (!pageLayoutID) {
                throw Error('field pageLayoutID is empty');
            }
            if (!translations) {
                throw Error('field translations is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.pageLayoutUpdateTranslationsEndpoint({
                    namespaceID, pageID, pageLayoutID,
                }) });
            cfg.data = {
                translations,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    pageLayoutUpdateTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.pageLayoutUpdateTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    pageLayoutUpdateTranslationsEndpoint(a) {
        const { namespaceID, pageID, pageLayoutID, } = a || {};
        return `/namespace/${namespaceID}/page/${pageID}/layout/${pageLayoutID}/translation`;
    }
    // List modules
    moduleList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, query, name, handle, limit, incTotal, pageCursor, labels, sort, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.moduleListEndpoint({
                    namespaceID,
                }) });
            cfg.params = {
                query,
                name,
                handle,
                limit,
                incTotal,
                pageCursor,
                labels,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    moduleListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.moduleList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    moduleListEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/module/`;
    }
    // Create module
    moduleCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, name, handle, config, meta, fields, labels, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            if (!meta) {
                throw Error('field meta is empty');
            }
            if (!fields) {
                throw Error('field fields is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.moduleCreateEndpoint({
                    namespaceID,
                }) });
            cfg.data = {
                name,
                handle,
                config,
                meta,
                fields,
                labels,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    moduleCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.moduleCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    moduleCreateEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/module/`;
    }
    // Read module
    moduleRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.moduleReadEndpoint({
                    namespaceID, moduleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    moduleReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.moduleRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    moduleReadEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}`;
    }
    // Update module
    moduleUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, name, handle, config, meta, fields, labels, updatedAt, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            if (!meta) {
                throw Error('field meta is empty');
            }
            if (!fields) {
                throw Error('field fields is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.moduleUpdateEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                name,
                handle,
                config,
                meta,
                fields,
                labels,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    moduleUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.moduleUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    moduleUpdateEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}`;
    }
    // Delete module
    moduleDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.moduleDeleteEndpoint({
                    namespaceID, moduleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    moduleDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.moduleDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    moduleDeleteEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}`;
    }
    // Fire compose:module trigger
    moduleTriggerScript(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, script, args, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!script) {
                throw Error('field script is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.moduleTriggerScriptEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                script,
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    moduleTriggerScriptCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.moduleTriggerScript(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    moduleTriggerScriptEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/trigger`;
    }
    // List moudle translation
    moduleListTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.moduleListTranslationsEndpoint({
                    namespaceID, moduleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    moduleListTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.moduleListTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    moduleListTranslationsEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/translation`;
    }
    // Update module translation
    moduleUpdateTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, translations, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!translations) {
                throw Error('field translations is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.moduleUpdateTranslationsEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                translations,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    moduleUpdateTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.moduleUpdateTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    moduleUpdateTranslationsEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/translation`;
    }
    // Generates report from module records
    recordReport(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, metrics, dimensions, filter, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!dimensions) {
                throw Error('field dimensions is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.recordReportEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.params = {
                metrics,
                dimensions,
                filter,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordReportCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordReport(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordReportEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/report`;
    }
    // List/read records from module section
    recordList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, summaries, query, meta, deleted, limit, incTotal, incPageNavigation, pageCursor, sort, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.recordListEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.params = {
                summaries,
                query,
                meta,
                deleted,
                limit,
                incTotal,
                incPageNavigation,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordListEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/`;
    }
    // Initiate record import session
    recordImportInit(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, upload, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!upload) {
                throw Error('field upload is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.recordImportInitEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                upload,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordImportInitCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordImportInit(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordImportInitEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/import`;
    }
    // Run record import
    recordImportRun(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, sessionID, fields, onError, multiValueDelimiter, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!sessionID) {
                throw Error('field sessionID is empty');
            }
            if (!fields) {
                throw Error('field fields is empty');
            }
            if (!onError) {
                throw Error('field onError is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.recordImportRunEndpoint({
                    namespaceID, moduleID, sessionID,
                }) });
            cfg.data = {
                fields,
                onError,
                multiValueDelimiter,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordImportRunCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordImportRun(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordImportRunEndpoint(a) {
        const { namespaceID, moduleID, sessionID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/import/${sessionID}`;
    }
    // Get import progress
    recordImportProgress(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, sessionID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!sessionID) {
                throw Error('field sessionID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.recordImportProgressEndpoint({
                    namespaceID, moduleID, sessionID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordImportProgressCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordImportProgress(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordImportProgressEndpoint(a) {
        const { namespaceID, moduleID, sessionID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/import/${sessionID}`;
    }
    // Exports records that match
    recordExport(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, filename, ext, filter, fields, timezone, multiValueDelimiter, wrapMultiValue, resolveRefs, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!ext) {
                throw Error('field ext is empty');
            }
            if (!fields) {
                throw Error('field fields is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.recordExportEndpoint({
                    namespaceID, moduleID, filename, ext,
                }) });
            cfg.params = {
                filter,
                fields,
                timezone,
                multiValueDelimiter,
                wrapMultiValue,
                resolveRefs,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordExportCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordExport(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordExportEndpoint(a) {
        const { namespaceID, moduleID, filename, ext, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/export${filename}.${ext}`;
    }
    // Executes server-side procedure over one or more module records
    recordExec(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, procedure, args, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!procedure) {
                throw Error('field procedure is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.recordExecEndpoint({
                    namespaceID, moduleID, procedure,
                }) });
            cfg.data = {
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordExecCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordExec(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordExecEndpoint(a) {
        const { namespaceID, moduleID, procedure, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/exec/${procedure}`;
    }
    // Create record in module section
    recordCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, values, ownedBy, records, meta, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.recordCreateEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                values,
                ownedBy,
                records,
                meta,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordCreateEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/`;
    }
    // Read records by ID from module section
    recordRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, recordID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!recordID) {
                throw Error('field recordID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.recordReadEndpoint({
                    namespaceID, moduleID, recordID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordReadEndpoint(a) {
        const { namespaceID, moduleID, recordID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/${recordID}`;
    }
    // Update records in module section
    recordUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, recordID, values, ownedBy, meta, records, updatedAt, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!recordID) {
                throw Error('field recordID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.recordUpdateEndpoint({
                    namespaceID, moduleID, recordID,
                }) });
            cfg.data = {
                values,
                ownedBy,
                meta,
                records,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordUpdateEndpoint(a) {
        const { namespaceID, moduleID, recordID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/${recordID}`;
    }
    // Partially update record values
    recordPatch(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, values, query, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.recordPatchEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                values,
                query,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordPatchCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordPatch(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordPatchEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/`;
    }
    // Delete record row from module section
    recordBulkDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, truncate, query, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.recordBulkDeleteEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                truncate,
                query,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordBulkDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordBulkDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordBulkDeleteEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/`;
    }
    // Delete record row from module section
    recordDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, recordID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!recordID) {
                throw Error('field recordID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.recordDeleteEndpoint({
                    namespaceID, moduleID, recordID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordDeleteEndpoint(a) {
        const { namespaceID, moduleID, recordID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/${recordID}`;
    }
    // Undelete soft-deleted record from module section
    recordUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, recordID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!recordID) {
                throw Error('field recordID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.recordUndeleteEndpoint({
                    namespaceID, moduleID, recordID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordUndeleteEndpoint(a) {
        const { namespaceID, moduleID, recordID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/${recordID}/undelete`;
    }
    // Undelete soft-deleted records from module section
    recordBulkUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, query, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.recordBulkUndeleteEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                query,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordBulkUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordBulkUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordBulkUndeleteEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/undelete`;
    }
    // Uploads attachment and validates it against record field requirements
    recordUpload(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, recordID, fieldName, upload, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!fieldName) {
                throw Error('field fieldName is empty');
            }
            if (!upload) {
                throw Error('field upload is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.recordUploadEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                recordID,
                fieldName,
                upload,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordUploadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordUpload(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordUploadEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/attachment`;
    }
    // Fire compose:record trigger
    recordTriggerScript(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, recordID, script, values, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!recordID) {
                throw Error('field recordID is empty');
            }
            if (!script) {
                throw Error('field script is empty');
            }
            if (!values) {
                throw Error('field values is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.recordTriggerScriptEndpoint({
                    namespaceID, moduleID, recordID,
                }) });
            cfg.data = {
                script,
                values,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordTriggerScriptCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordTriggerScript(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordTriggerScriptEndpoint(a) {
        const { namespaceID, moduleID, recordID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/${recordID}/trigger`;
    }
    // Fire compose:record trigger
    recordTriggerScriptOnList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, script, args, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!script) {
                throw Error('field script is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.recordTriggerScriptOnListEndpoint({
                    namespaceID, moduleID,
                }) });
            cfg.data = {
                script,
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordTriggerScriptOnListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordTriggerScriptOnList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordTriggerScriptOnListEndpoint(a) {
        const { namespaceID, moduleID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/trigger`;
    }
    // List record revisions
    recordRevisions(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, moduleID, recordID, sort, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!recordID) {
                throw Error('field recordID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.recordRevisionsEndpoint({
                    namespaceID, moduleID, recordID,
                }) });
            cfg.params = {
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    recordRevisionsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.recordRevisions(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    recordRevisionsEndpoint(a) {
        const { namespaceID, moduleID, recordID, } = a || {};
        return `/namespace/${namespaceID}/module/${moduleID}/record/${recordID}/revisions`;
    }
    // List records for data privacy
    dataPrivacyRecordList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sensitivityLevelID, connectionID, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dataPrivacyRecordListEndpoint() });
            cfg.params = {
                sensitivityLevelID,
                connectionID,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    dataPrivacyRecordListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dataPrivacyRecordList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dataPrivacyRecordListEndpoint() {
        return '/data-privacy/record';
    }
    // List modules
    dataPrivacyModuleList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { connectionID, limit, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.dataPrivacyModuleListEndpoint() });
            cfg.params = {
                connectionID,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    dataPrivacyModuleListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.dataPrivacyModuleList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    dataPrivacyModuleListEndpoint() {
        return '/data-privacy/module';
    }
    // List/read charts
    chartList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, query, handle, labels, limit, incTotal, pageCursor, sort, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.chartListEndpoint({
                    namespaceID,
                }) });
            cfg.params = {
                query,
                handle,
                labels,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    chartListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.chartList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    chartListEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/chart/`;
    }
    // List/read charts
    chartCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, config, name, handle, labels, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!config) {
                throw Error('field config is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.chartCreateEndpoint({
                    namespaceID,
                }) });
            cfg.data = {
                config,
                name,
                handle,
                labels,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    chartCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.chartCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    chartCreateEndpoint(a) {
        const { namespaceID, } = a || {};
        return `/namespace/${namespaceID}/chart/`;
    }
    // Read charts by ID
    chartRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, chartID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!chartID) {
                throw Error('field chartID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.chartReadEndpoint({
                    namespaceID, chartID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    chartReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.chartRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    chartReadEndpoint(a) {
        const { namespaceID, chartID, } = a || {};
        return `/namespace/${namespaceID}/chart/${chartID}`;
    }
    // Add/update charts
    chartUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, chartID, config, name, handle, labels, updatedAt, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!chartID) {
                throw Error('field chartID is empty');
            }
            if (!config) {
                throw Error('field config is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.chartUpdateEndpoint({
                    namespaceID, chartID,
                }) });
            cfg.data = {
                config,
                name,
                handle,
                labels,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    chartUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.chartUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    chartUpdateEndpoint(a) {
        const { namespaceID, chartID, } = a || {};
        return `/namespace/${namespaceID}/chart/${chartID}`;
    }
    // Delete chart
    chartDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, chartID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!chartID) {
                throw Error('field chartID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.chartDeleteEndpoint({
                    namespaceID, chartID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    chartDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.chartDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    chartDeleteEndpoint(a) {
        const { namespaceID, chartID, } = a || {};
        return `/namespace/${namespaceID}/chart/${chartID}`;
    }
    // List chart translation
    chartListTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, chartID, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!chartID) {
                throw Error('field chartID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.chartListTranslationsEndpoint({
                    namespaceID, chartID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    chartListTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.chartListTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    chartListTranslationsEndpoint(a) {
        const { namespaceID, chartID, } = a || {};
        return `/namespace/${namespaceID}/chart/${chartID}/translation`;
    }
    // Update chart translation
    chartUpdateTranslations(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { namespaceID, chartID, translations, } = a || {};
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!chartID) {
                throw Error('field chartID is empty');
            }
            if (!translations) {
                throw Error('field translations is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.chartUpdateTranslationsEndpoint({
                    namespaceID, chartID,
                }) });
            cfg.data = {
                translations,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    chartUpdateTranslationsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.chartUpdateTranslations(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    chartUpdateTranslationsEndpoint(a) {
        const { namespaceID, chartID, } = a || {};
        return `/namespace/${namespaceID}/chart/${chartID}/translation`;
    }
    // Send email from the Compose
    notificationEmailSend(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { to, cc, replyTo, subject, content, remoteAttachments, } = a || {};
            if (!to) {
                throw Error('field to is empty');
            }
            if (!content) {
                throw Error('field content is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.notificationEmailSendEndpoint() });
            cfg.data = {
                to,
                cc,
                replyTo,
                subject,
                content,
                remoteAttachments,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    notificationEmailSendCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.notificationEmailSend(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    notificationEmailSendEndpoint() {
        return '/notification/email';
    }
    // List, filter all page attachments
    attachmentList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, namespaceID, sign, userID, pageID, moduleID, recordID, fieldName, limit, pageCursor, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.attachmentListEndpoint({
                    kind, namespaceID,
                }) });
            cfg.params = {
                sign,
                userID,
                pageID,
                moduleID,
                recordID,
                fieldName,
                limit,
                pageCursor,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    attachmentListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.attachmentList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    attachmentListEndpoint(a) {
        const { kind, namespaceID, } = a || {};
        return `/namespace/${namespaceID}/attachment/${kind}/`;
    }
    // Attachment details
    attachmentRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, namespaceID, attachmentID, sign, userID, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!attachmentID) {
                throw Error('field attachmentID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.attachmentReadEndpoint({
                    kind, namespaceID, attachmentID,
                }) });
            cfg.params = {
                sign,
                userID,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    attachmentReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.attachmentRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    attachmentReadEndpoint(a) {
        const { kind, namespaceID, attachmentID, } = a || {};
        return `/namespace/${namespaceID}/attachment/${kind}/${attachmentID}`;
    }
    // Delete attachment
    attachmentDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, namespaceID, attachmentID, sign, userID, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!attachmentID) {
                throw Error('field attachmentID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.attachmentDeleteEndpoint({
                    kind, namespaceID, attachmentID,
                }) });
            cfg.params = {
                sign,
                userID,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    attachmentDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.attachmentDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    attachmentDeleteEndpoint(a) {
        const { kind, namespaceID, attachmentID, } = a || {};
        return `/namespace/${namespaceID}/attachment/${kind}/${attachmentID}`;
    }
    // Serves attached file
    attachmentOriginal(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, namespaceID, attachmentID, name, sign, userID, download, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!attachmentID) {
                throw Error('field attachmentID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.attachmentOriginalEndpoint({
                    kind, namespaceID, attachmentID, name,
                }) });
            cfg.params = {
                sign,
                userID,
                download,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    attachmentOriginalCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.attachmentOriginal(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    attachmentOriginalEndpoint(a) {
        const { kind, namespaceID, attachmentID, name, } = a || {};
        return `/namespace/${namespaceID}/attachment/${kind}/${attachmentID}/original/${name}`;
    }
    // Serves preview of an attached file
    attachmentPreview(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { kind, namespaceID, attachmentID, ext, sign, userID, } = a || {};
            if (!kind) {
                throw Error('field kind is empty');
            }
            if (!namespaceID) {
                throw Error('field namespaceID is empty');
            }
            if (!attachmentID) {
                throw Error('field attachmentID is empty');
            }
            if (!ext) {
                throw Error('field ext is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.attachmentPreviewEndpoint({
                    kind, namespaceID, attachmentID, ext,
                }) });
            cfg.params = {
                sign,
                userID,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    attachmentPreviewCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.attachmentPreview(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    attachmentPreviewEndpoint(a) {
        const { kind, namespaceID, attachmentID, ext, } = a || {};
        return `/namespace/${namespaceID}/attachment/${kind}/${attachmentID}/preview.${ext}`;
    }
    // Retrieve defined permissions
    permissionsList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    permissionsListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsListEndpoint() {
        return '/permissions/';
    }
    // Effective rules for current user
    permissionsEffective(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resource, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsEffectiveEndpoint() });
            cfg.params = {
                resource,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    permissionsEffectiveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsEffective(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsEffectiveEndpoint() {
        return '/permissions/effective';
    }
    // Evaluate rules for given user/role combo
    permissionsTrace(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resource, userID, roleID, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsTraceEndpoint() });
            cfg.params = {
                resource,
                userID,
                roleID,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    permissionsTraceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsTrace(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsTraceEndpoint() {
        return '/permissions/trace';
    }
    // Retrieve role permissions
    permissionsRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, resource, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsReadEndpoint({
                    roleID,
                }) });
            cfg.params = {
                resource,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    permissionsReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsReadEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // Remove all defined role permissions
    permissionsDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.permissionsDeleteEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    permissionsDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsDeleteEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // Update permission settings
    permissionsUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, rules, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!rules) {
                throw Error('field rules is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.permissionsUpdateEndpoint({
                    roleID,
                }) });
            cfg.data = {
                rules,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    permissionsUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsUpdateEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // List all available automation scripts for compose resources
    automationList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resourceTypePrefixes, resourceTypes, eventTypes, excludeInvalid, excludeClientScripts, excludeServerScripts, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.automationListEndpoint() });
            cfg.params = {
                resourceTypePrefixes,
                resourceTypes,
                eventTypes,
                excludeInvalid,
                excludeClientScripts,
                excludeServerScripts,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    automationListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.automationList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    automationListEndpoint() {
        return '/automation/';
    }
    // Serves client scripts bundle
    automationBundle(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { bundle, type, ext, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.automationBundleEndpoint({
                    bundle, type, ext,
                }) });
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    automationBundleCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.automationBundle(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    automationBundleEndpoint(a) {
        const { bundle, type, ext, } = a || {};
        return `/automation/${bundle}-${type}.${ext}`;
    }
    // Triggers execution of a specific script on a system service level
    automationTriggerScript(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { script, args, } = a || {};
            if (!script) {
                throw Error('field script is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.automationTriggerScriptEndpoint() });
            cfg.data = {
                script,
                args,
            };
            return this.api().request(cfg).then(result => stdResolve$3(result));
        });
    }
    automationTriggerScriptCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.automationTriggerScript(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    automationTriggerScriptEndpoint() {
        return '/automation/trigger';
    }
}

/* eslint-disable padded-blocks */
function stdResolve$2(response) {
    if (response.data.error) {
        return Promise.reject(response.data.error);
    }
    else {
        return response.data.response;
    }
}
class Federation {
    constructor({ baseURL, headers, accessTokenFn }) {
        this.headers = {};
        this.baseURL = baseURL;
        this.accessTokenFn = accessTokenFn;
        this.headers = {
            /**
             * All we send is JSON
             */
            'Content-Type': 'application/json',
        };
        this.setHeaders(headers);
    }
    setAccessTokenFn(fn) {
        this.accessTokenFn = fn;
        return this;
    }
    setHeaders(headers) {
        if (typeof headers === 'object') {
            this.headers = headers;
        }
        return this;
    }
    setHeader(name, value) {
        if (value === undefined) {
            delete this.headers[name];
        }
        else {
            this.headers[name] = value;
        }
        return this;
    }
    api() {
        const headers = Object.assign({}, this.headers);
        const accessToken = this.accessTokenFn ? this.accessTokenFn() : undefined;
        if (accessToken) {
            headers.Authorization = 'Bearer ' + accessToken;
        }
        return axios.create({
            withCredentials: true,
            baseURL: this.baseURL,
            headers,
        });
    }
    // Initialize the handshake step with node B
    nodeHandshakeInitialize(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, pairToken, sharedNodeID, authToken, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!pairToken) {
                throw Error('field pairToken is empty');
            }
            if (!sharedNodeID) {
                throw Error('field sharedNodeID is empty');
            }
            if (!authToken) {
                throw Error('field authToken is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.nodeHandshakeInitializeEndpoint({
                    nodeID,
                }) });
            cfg.data = {
                pairToken,
                sharedNodeID,
                authToken,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeHandshakeInitializeCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeHandshakeInitialize(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeHandshakeInitializeEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/handshake`;
    }
    // Search federated nodes
    nodeSearch(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { query, status, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.nodeSearchEndpoint() });
            cfg.params = {
                query,
                status,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeSearchCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeSearch(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeSearchEndpoint() {
        return '/nodes/';
    }
    // Create a new federation node
    nodeCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { baseURL, name, contact, pairingURI, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.nodeCreateEndpoint() });
            cfg.data = {
                baseURL,
                name,
                contact,
                pairingURI,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeCreateEndpoint() {
        return '/nodes/';
    }
    // Read a federation node
    nodeRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.nodeReadEndpoint({
                    nodeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeReadEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}`;
    }
    // Creates new sharable federation URI
    nodeGenerateUri(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.nodeGenerateUriEndpoint({
                    nodeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeGenerateUriCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeGenerateUri(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeGenerateUriEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/uri`;
    }
    // Updates existing node
    nodeUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, name, contact, baseURL, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.nodeUpdateEndpoint({
                    nodeID,
                }) });
            cfg.data = {
                name,
                contact,
                baseURL,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeUpdateEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}`;
    }
    // Deletes node
    nodeDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.nodeDeleteEndpoint({
                    nodeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeDeleteEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}`;
    }
    // Undeletes a node
    nodeUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.nodeUndeleteEndpoint({
                    nodeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeUndeleteEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/undelete`;
    }
    // Initialize the pairing process between the two nodes
    nodePair(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.nodePairEndpoint({
                    nodeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodePairCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodePair(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodePairEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/pair`;
    }
    // Confirm the requested handshake
    nodeHandshakeConfirm(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.nodeHandshakeConfirmEndpoint({
                    nodeID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeHandshakeConfirmCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeHandshakeConfirm(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeHandshakeConfirmEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/handshake-confirm`;
    }
    // Complete the handshake
    nodeHandshakeComplete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, authToken, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!authToken) {
                throw Error('field authToken is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.nodeHandshakeCompleteEndpoint({
                    nodeID,
                }) });
            cfg.data = {
                authToken,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    nodeHandshakeCompleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.nodeHandshakeComplete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    nodeHandshakeCompleteEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/handshake-complete`;
    }
    // Exposed settings for module
    manageStructureReadExposed(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, moduleID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.manageStructureReadExposedEndpoint({
                    nodeID, moduleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    manageStructureReadExposedCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.manageStructureReadExposed(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    manageStructureReadExposedEndpoint(a) {
        const { nodeID, moduleID, } = a || {};
        return `/nodes/${nodeID}/modules/${moduleID}/exposed`;
    }
    // Add module to federation
    manageStructureCreateExposed(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, composeModuleID, composeNamespaceID, name, handle, fields, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!composeModuleID) {
                throw Error('field composeModuleID is empty');
            }
            if (!composeNamespaceID) {
                throw Error('field composeNamespaceID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            if (!handle) {
                throw Error('field handle is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.manageStructureCreateExposedEndpoint({
                    nodeID,
                }) });
            cfg.data = {
                composeModuleID,
                composeNamespaceID,
                name,
                handle,
                fields,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    manageStructureCreateExposedCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.manageStructureCreateExposed(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    manageStructureCreateExposedEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/modules/`;
    }
    // Update already exposed module
    manageStructureUpdateExposed(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, moduleID, composeModuleID, composeNamespaceID, name, handle, fields, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!composeModuleID) {
                throw Error('field composeModuleID is empty');
            }
            if (!composeNamespaceID) {
                throw Error('field composeNamespaceID is empty');
            }
            if (!name) {
                throw Error('field name is empty');
            }
            if (!handle) {
                throw Error('field handle is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.manageStructureUpdateExposedEndpoint({
                    nodeID, moduleID,
                }) });
            cfg.data = {
                composeModuleID,
                composeNamespaceID,
                name,
                handle,
                fields,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    manageStructureUpdateExposedCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.manageStructureUpdateExposed(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    manageStructureUpdateExposedEndpoint(a) {
        const { nodeID, moduleID, } = a || {};
        return `/nodes/${nodeID}/modules/${moduleID}/exposed`;
    }
    // Remove from federation
    manageStructureRemoveExposed(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, moduleID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.manageStructureRemoveExposedEndpoint({
                    nodeID, moduleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    manageStructureRemoveExposedCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.manageStructureRemoveExposed(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    manageStructureRemoveExposedEndpoint(a) {
        const { nodeID, moduleID, } = a || {};
        return `/nodes/${nodeID}/modules/${moduleID}/exposed`;
    }
    // Shared settings for module
    manageStructureReadShared(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, moduleID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.manageStructureReadSharedEndpoint({
                    nodeID, moduleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    manageStructureReadSharedCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.manageStructureReadShared(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    manageStructureReadSharedEndpoint(a) {
        const { nodeID, moduleID, } = a || {};
        return `/nodes/${nodeID}/modules/${moduleID}/shared`;
    }
    // Add fields mappings to federated module
    manageStructureCreateMappings(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, moduleID, composeModuleID, composeNamespaceID, fields, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            if (!composeModuleID) {
                throw Error('field composeModuleID is empty');
            }
            if (!composeNamespaceID) {
                throw Error('field composeNamespaceID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.manageStructureCreateMappingsEndpoint({
                    nodeID, moduleID,
                }) });
            cfg.data = {
                composeModuleID,
                composeNamespaceID,
                fields,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    manageStructureCreateMappingsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.manageStructureCreateMappings(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    manageStructureCreateMappingsEndpoint(a) {
        const { nodeID, moduleID, } = a || {};
        return `/nodes/${nodeID}/modules/${moduleID}/mapped`;
    }
    // Fields mappings for module
    manageStructureReadMappings(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, moduleID, composeModuleID, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.manageStructureReadMappingsEndpoint({
                    nodeID, moduleID,
                }) });
            cfg.params = {
                composeModuleID,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    manageStructureReadMappingsCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.manageStructureReadMappings(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    manageStructureReadMappingsEndpoint(a) {
        const { nodeID, moduleID, } = a || {};
        return `/nodes/${nodeID}/modules/${moduleID}/mapped`;
    }
    // List of shared/exposed/mapped modules
    manageStructureListAll(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, shared, exposed, mapped, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.manageStructureListAllEndpoint({
                    nodeID,
                }) });
            cfg.params = {
                shared,
                exposed,
                mapped,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    manageStructureListAllCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.manageStructureListAll(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    manageStructureListAllEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/modules/`;
    }
    // List all exposed modules changes
    syncStructureReadExposedInternal(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, lastSync, query, limit, pageCursor, sort, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.syncStructureReadExposedInternalEndpoint({
                    nodeID,
                }) });
            cfg.params = {
                lastSync,
                query,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    syncStructureReadExposedInternalCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.syncStructureReadExposedInternal(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    syncStructureReadExposedInternalEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/modules/exposed/`;
    }
    // List all exposed modules changes in activity streams format
    syncStructureReadExposedSocial(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, lastSync, query, limit, pageCursor, sort, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.syncStructureReadExposedSocialEndpoint({
                    nodeID,
                }) });
            cfg.params = {
                lastSync,
                query,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    syncStructureReadExposedSocialCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.syncStructureReadExposedSocial(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    syncStructureReadExposedSocialEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/modules/exposed/activity-stream`;
    }
    // List all record changes
    syncDataReadExposedAll(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, lastSync, query, limit, pageCursor, sort, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.syncDataReadExposedAllEndpoint({
                    nodeID,
                }) });
            cfg.params = {
                lastSync,
                query,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    syncDataReadExposedAllCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.syncDataReadExposedAll(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    syncDataReadExposedAllEndpoint(a) {
        const { nodeID, } = a || {};
        return `/nodes/${nodeID}/modules/exposed/records/`;
    }
    // List all records per module
    syncDataReadExposedInternal(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, moduleID, lastSync, query, limit, pageCursor, sort, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.syncDataReadExposedInternalEndpoint({
                    nodeID, moduleID,
                }) });
            cfg.params = {
                lastSync,
                query,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    syncDataReadExposedInternalCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.syncDataReadExposedInternal(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    syncDataReadExposedInternalEndpoint(a) {
        const { nodeID, moduleID, } = a || {};
        return `/nodes/${nodeID}/modules/${moduleID}/records/`;
    }
    // List all records per module in activitystreams format
    syncDataReadExposedSocial(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { nodeID, moduleID, lastSync, query, limit, pageCursor, sort, } = a || {};
            if (!nodeID) {
                throw Error('field nodeID is empty');
            }
            if (!moduleID) {
                throw Error('field moduleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.syncDataReadExposedSocialEndpoint({
                    nodeID, moduleID,
                }) });
            cfg.params = {
                lastSync,
                query,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    syncDataReadExposedSocialCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.syncDataReadExposedSocial(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    syncDataReadExposedSocialEndpoint(a) {
        const { nodeID, moduleID, } = a || {};
        return `/nodes/${nodeID}/modules/${moduleID}/records/activity-stream/`;
    }
    // Retrieve defined permissions
    permissionsList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    permissionsListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsListEndpoint() {
        return '/permissions/';
    }
    // Effective rules for current user
    permissionsEffective(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resource, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsEffectiveEndpoint() });
            cfg.params = {
                resource,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    permissionsEffectiveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsEffective(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsEffectiveEndpoint() {
        return '/permissions/effective';
    }
    // Evaluate rules for given user/role combo
    permissionsTrace(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resource, userID, roleID, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsTraceEndpoint() });
            cfg.params = {
                resource,
                userID,
                roleID,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    permissionsTraceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsTrace(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsTraceEndpoint() {
        return '/permissions/trace';
    }
    // Retrieve role permissions
    permissionsRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, resource, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsReadEndpoint({
                    roleID,
                }) });
            cfg.params = {
                resource,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    permissionsReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsReadEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // Remove all defined role permissions
    permissionsDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.permissionsDeleteEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    permissionsDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsDeleteEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // Update permission settings
    permissionsUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, rules, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!rules) {
                throw Error('field rules is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.permissionsUpdateEndpoint({
                    roleID,
                }) });
            cfg.data = {
                rules,
            };
            return this.api().request(cfg).then(result => stdResolve$2(result));
        });
    }
    permissionsUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsUpdateEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
}

/* eslint-disable padded-blocks */
function stdResolve$1(response) {
    if (response.data.error) {
        return Promise.reject(response.data.error);
    }
    else {
        return response.data.response;
    }
}
class Automation {
    constructor({ baseURL, headers, accessTokenFn }) {
        this.headers = {};
        this.baseURL = baseURL;
        this.accessTokenFn = accessTokenFn;
        this.headers = {
            /**
             * All we send is JSON
             */
            'Content-Type': 'application/json',
        };
        this.setHeaders(headers);
    }
    setAccessTokenFn(fn) {
        this.accessTokenFn = fn;
        return this;
    }
    setHeaders(headers) {
        if (typeof headers === 'object') {
            this.headers = headers;
        }
        return this;
    }
    setHeader(name, value) {
        if (value === undefined) {
            delete this.headers[name];
        }
        else {
            this.headers[name] = value;
        }
        return this;
    }
    api() {
        const headers = Object.assign({}, this.headers);
        const accessToken = this.accessTokenFn ? this.accessTokenFn() : undefined;
        if (accessToken) {
            headers.Authorization = 'Bearer ' + accessToken;
        }
        return axios.create({
            withCredentials: true,
            baseURL: this.baseURL,
            headers,
        });
    }
    // List workflows
    workflowList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { workflowID, query, deleted, disabled, subWorkflow, labels, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.workflowListEndpoint() });
            cfg.params = {
                workflowID,
                query,
                deleted,
                disabled,
                subWorkflow,
                labels,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    workflowListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.workflowList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    workflowListEndpoint() {
        return '/workflows/';
    }
    // Create workflow
    workflowCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { handle, labels, meta, enabled, trace, keepSessions, scope, steps, paths, runAs, ownedBy, } = a || {};
            if (!runAs) {
                throw Error('field runAs is empty');
            }
            if (!ownedBy) {
                throw Error('field ownedBy is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.workflowCreateEndpoint() });
            cfg.data = {
                handle,
                labels,
                meta,
                enabled,
                trace,
                keepSessions,
                scope,
                steps,
                paths,
                runAs,
                ownedBy,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    workflowCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.workflowCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    workflowCreateEndpoint() {
        return '/workflows/';
    }
    // Update triger details
    workflowUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { workflowID, handle, labels, meta, enabled, trace, keepSessions, scope, steps, paths, runAs, ownedBy, updatedAt, } = a || {};
            if (!workflowID) {
                throw Error('field workflowID is empty');
            }
            if (!runAs) {
                throw Error('field runAs is empty');
            }
            if (!ownedBy) {
                throw Error('field ownedBy is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.workflowUpdateEndpoint({
                    workflowID,
                }) });
            cfg.data = {
                handle,
                labels,
                meta,
                enabled,
                trace,
                keepSessions,
                scope,
                steps,
                paths,
                runAs,
                ownedBy,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    workflowUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.workflowUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    workflowUpdateEndpoint(a) {
        const { workflowID, } = a || {};
        return `/workflows/${workflowID}`;
    }
    // Read workflow details
    workflowRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { workflowID, } = a || {};
            if (!workflowID) {
                throw Error('field workflowID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.workflowReadEndpoint({
                    workflowID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    workflowReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.workflowRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    workflowReadEndpoint(a) {
        const { workflowID, } = a || {};
        return `/workflows/${workflowID}`;
    }
    // Remove workflow
    workflowDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { workflowID, } = a || {};
            if (!workflowID) {
                throw Error('field workflowID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.workflowDeleteEndpoint({
                    workflowID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    workflowDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.workflowDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    workflowDeleteEndpoint(a) {
        const { workflowID, } = a || {};
        return `/workflows/${workflowID}`;
    }
    // Undelete workflow
    workflowUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { workflowID, } = a || {};
            if (!workflowID) {
                throw Error('field workflowID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.workflowUndeleteEndpoint({
                    workflowID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    workflowUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.workflowUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    workflowUndeleteEndpoint(a) {
        const { workflowID, } = a || {};
        return `/workflows/${workflowID}/undelete`;
    }
    // Test workflow details
    workflowTest(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { workflowID, scope, runAs, } = a || {};
            if (!workflowID) {
                throw Error('field workflowID is empty');
            }
            if (!runAs) {
                throw Error('field runAs is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.workflowTestEndpoint({
                    workflowID,
                }) });
            cfg.data = {
                scope,
                runAs,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    workflowTestCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.workflowTest(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    workflowTestEndpoint(a) {
        const { workflowID, } = a || {};
        return `/workflows/${workflowID}/test`;
    }
    // Executes workflow on a specific step (must be orphan step and connected to &#x27;onManual&#x27; trigger)
    workflowExec(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { workflowID, stepID, input, trace, wait, async, } = a || {};
            if (!workflowID) {
                throw Error('field workflowID is empty');
            }
            if (!stepID) {
                throw Error('field stepID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.workflowExecEndpoint({
                    workflowID,
                }) });
            cfg.data = {
                stepID,
                input,
                trace,
                wait,
                async,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    workflowExecCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.workflowExec(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    workflowExecEndpoint(a) {
        const { workflowID, } = a || {};
        return `/workflows/${workflowID}/exec`;
    }
    // List triggers
    triggerList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { triggerID, workflowID, deleted, disabled, eventType, resourceType, query, labels, limit, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.triggerListEndpoint() });
            cfg.params = {
                triggerID,
                workflowID,
                deleted,
                disabled,
                eventType,
                resourceType,
                query,
                labels,
                limit,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    triggerListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.triggerList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    triggerListEndpoint() {
        return '/triggers/';
    }
    // Create trigger
    triggerCreate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { eventType, resourceType, enabled, workflowID, workflowStepID, input, labels, meta, constraints, ownedBy, } = a || {};
            if (!eventType) {
                throw Error('field eventType is empty');
            }
            if (!resourceType) {
                throw Error('field resourceType is empty');
            }
            if (!workflowID) {
                throw Error('field workflowID is empty');
            }
            if (!workflowStepID) {
                throw Error('field workflowStepID is empty');
            }
            if (!ownedBy) {
                throw Error('field ownedBy is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.triggerCreateEndpoint() });
            cfg.data = {
                eventType,
                resourceType,
                enabled,
                workflowID,
                workflowStepID,
                input,
                labels,
                meta,
                constraints,
                ownedBy,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    triggerCreateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.triggerCreate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    triggerCreateEndpoint() {
        return '/triggers/';
    }
    // Update trigger details
    triggerUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { triggerID, eventType, resourceType, enabled, workflowID, workflowStepID, input, labels, meta, constraints, ownedBy, updatedAt, } = a || {};
            if (!triggerID) {
                throw Error('field triggerID is empty');
            }
            if (!eventType) {
                throw Error('field eventType is empty');
            }
            if (!resourceType) {
                throw Error('field resourceType is empty');
            }
            if (!workflowID) {
                throw Error('field workflowID is empty');
            }
            if (!workflowStepID) {
                throw Error('field workflowStepID is empty');
            }
            if (!ownedBy) {
                throw Error('field ownedBy is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'put', url: this.triggerUpdateEndpoint({
                    triggerID,
                }) });
            cfg.data = {
                eventType,
                resourceType,
                enabled,
                workflowID,
                workflowStepID,
                input,
                labels,
                meta,
                constraints,
                ownedBy,
                updatedAt,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    triggerUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.triggerUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    triggerUpdateEndpoint(a) {
        const { triggerID, } = a || {};
        return `/triggers/${triggerID}`;
    }
    // Read trigger details
    triggerRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { triggerID, } = a || {};
            if (!triggerID) {
                throw Error('field triggerID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.triggerReadEndpoint({
                    triggerID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    triggerReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.triggerRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    triggerReadEndpoint(a) {
        const { triggerID, } = a || {};
        return `/triggers/${triggerID}`;
    }
    // Remove trigger
    triggerDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { triggerID, } = a || {};
            if (!triggerID) {
                throw Error('field triggerID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.triggerDeleteEndpoint({
                    triggerID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    triggerDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.triggerDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    triggerDeleteEndpoint(a) {
        const { triggerID, } = a || {};
        return `/triggers/${triggerID}`;
    }
    // Undelete trigger
    triggerUndelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { triggerID, } = a || {};
            if (!triggerID) {
                throw Error('field triggerID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.triggerUndeleteEndpoint({
                    triggerID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    triggerUndeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.triggerUndelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    triggerUndeleteEndpoint(a) {
        const { triggerID, } = a || {};
        return `/triggers/${triggerID}/undelete`;
    }
    // List sessions
    sessionList(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sessionID, workflowID, createdBy, completed, status, eventType, resourceType, limit, incTotal, pageCursor, sort, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.sessionListEndpoint() });
            cfg.params = {
                sessionID,
                workflowID,
                createdBy,
                completed,
                status,
                eventType,
                resourceType,
                limit,
                incTotal,
                pageCursor,
                sort,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    sessionListCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.sessionList(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    sessionListEndpoint() {
        return '/sessions/';
    }
    // Read session details
    sessionRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sessionID, } = a || {};
            if (!sessionID) {
                throw Error('field sessionID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.sessionReadEndpoint({
                    sessionID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    sessionReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.sessionRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    sessionReadEndpoint(a) {
        const { sessionID, } = a || {};
        return `/sessions/${sessionID}`;
    }
    // Cancel session
    sessionCancel(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sessionID, } = a || {};
            if (!sessionID) {
                throw Error('field sessionID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.sessionCancelEndpoint({
                    sessionID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    sessionCancelCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.sessionCancel(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    sessionCancelEndpoint(a) {
        const { sessionID, } = a || {};
        return `/sessions/${sessionID}/cancel`;
    }
    // Returns pending prompts from all sessions
    sessionListPrompts() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.sessionListPromptsEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    sessionListPromptsCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.sessionListPrompts(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    sessionListPromptsEndpoint() {
        return '/sessions/prompts';
    }
    // Resume session
    sessionResumeState(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { sessionID, stateID, input, } = a || {};
            if (!sessionID) {
                throw Error('field sessionID is empty');
            }
            if (!stateID) {
                throw Error('field stateID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'post', url: this.sessionResumeStateEndpoint({
                    sessionID, stateID,
                }) });
            cfg.data = {
                input,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    sessionResumeStateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.sessionResumeState(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    sessionResumeStateEndpoint(a) {
        const { sessionID, stateID, } = a || {};
        return `/sessions/${sessionID}/state/${stateID}`;
    }
    // Available workflow functions
    functionList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.functionListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    functionListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.functionList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    functionListEndpoint() {
        return '/functions/';
    }
    // Available workflow types
    typeList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.typeListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    typeListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.typeList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    typeListEndpoint() {
        return '/types/';
    }
    // Available workflow types
    eventTypesList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.eventTypesListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    eventTypesListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.eventTypesList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    eventTypesListEndpoint() {
        return '/event-types/';
    }
    // Retrieve defined permissions
    permissionsList() {
        return __awaiter(this, arguments, void 0, function* (extra = {}) {
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsListEndpoint() });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    permissionsListCancellable(extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsList(options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsListEndpoint() {
        return '/permissions/';
    }
    // Effective rules for current user
    permissionsEffective(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resource, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsEffectiveEndpoint() });
            cfg.params = {
                resource,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    permissionsEffectiveCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsEffective(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsEffectiveEndpoint() {
        return '/permissions/effective';
    }
    // Evaluate rules for given user/role combo
    permissionsTrace(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { resource, userID, roleID, } = a || {};
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsTraceEndpoint() });
            cfg.params = {
                resource,
                userID,
                roleID,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    permissionsTraceCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsTrace(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsTraceEndpoint() {
        return '/permissions/trace';
    }
    // Retrieve role permissions
    permissionsRead(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, resource, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: this.permissionsReadEndpoint({
                    roleID,
                }) });
            cfg.params = {
                resource,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    permissionsReadCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsRead(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsReadEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // Remove all defined role permissions
    permissionsDelete(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'delete', url: this.permissionsDeleteEndpoint({
                    roleID,
                }) });
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    permissionsDeleteCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsDelete(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsDeleteEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
    // Update permission settings
    permissionsUpdate(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { roleID, rules, } = a || {};
            if (!roleID) {
                throw Error('field roleID is empty');
            }
            if (!rules) {
                throw Error('field rules is empty');
            }
            const cfg = Object.assign(Object.assign({}, extra), { method: 'patch', url: this.permissionsUpdateEndpoint({
                    roleID,
                }) });
            cfg.data = {
                rules,
            };
            return this.api().request(cfg).then(result => stdResolve$1(result));
        });
    }
    permissionsUpdateCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.permissionsUpdate(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
    permissionsUpdateEndpoint(a) {
        const { roleID, } = a || {};
        return `/permissions/${roleID}/rules`;
    }
}

function stdResolve(response) {
    if (response.data.error) {
        return Promise.reject(response.data.error);
    }
    else {
        return response.data.response;
    }
}
class Discovery {
    constructor({ baseURL, headers, accessTokenFn }) {
        this.headers = {};
        this.baseURL = baseURL;
        this.accessTokenFn = accessTokenFn;
        this.headers = {
            'Content-Type': 'application/json',
        };
        this.setHeaders(headers);
    }
    setAccessTokenFn(fn) {
        this.accessTokenFn = fn;
        return this;
    }
    setHeaders(headers) {
        if (typeof headers === 'object') {
            this.headers = headers;
        }
        return this;
    }
    setHeader(name, value) {
        if (value === undefined) {
            delete this.headers[name];
        }
        else {
            this.headers[name] = value;
        }
        return this;
    }
    api() {
        const headers = Object.assign({}, this.headers);
        const accessToken = this.accessTokenFn ? this.accessTokenFn() : undefined;
        if (accessToken) {
            headers.Authorization = 'Bearer ' + accessToken;
        }
        return axios.create({
            withCredentials: true,
            baseURL: this.baseURL,
            headers,
        });
    }
    query(a_1) {
        return __awaiter(this, arguments, void 0, function* (a, extra = {}) {
            const { query = '', from, size, resourceTypes, } = a || {};
            const params = new URLSearchParams();
            if (resourceTypes && Array.isArray(resourceTypes)) {
                resourceTypes.forEach(t => params.append('resourceTypes', t));
            }
            if (from)
                params.append('from', from.toString());
            if (size)
                params.append('size', size.toString());
            const cfg = Object.assign(Object.assign({}, extra), { method: 'get', url: `/?q=${query}`, params });
            return this.api().request(cfg).then(result => stdResolve(result));
        });
    }
    queryCancellable(a, extra = {}) {
        const cancelTokenSource = axios.CancelToken.source();
        const options = Object.assign(Object.assign({}, extra), { cancelToken: cancelTokenSource.token });
        return {
            response: () => this.query(a, options),
            cancel: () => {
                cancelTokenSource.cancel();
            },
        };
    }
}

var index$2 = /*#__PURE__*/Object.freeze({
  __proto__: null,
  Automation: Automation,
  Compose: Compose,
  Discovery: Discovery,
  Federation: Federation,
  System: System
});

/**
 * Handles script execution context
 *
 * Context accepts pre-assembled *API props or it construct them fly from passed config
 *
 * Naming convention for properties:
 *  - Corteza classes, high-level helpers, API clients are upper cased
 *  - low-level helpers are lower cased
 *  - simple scalar are lower cased
 *  - $authUser is the only one prefixed with the dollar sign for historical reasons
 */
class Ctx {
    constructor(args, logger, a) {
        this.args = args;
        this.logger = logger;
        if (a) {
            Object.assign(this, a);
        }
    }
    /**
     * Alias for log, to make developer's life easier <3
     */
    get console() {
        return this.logger;
    }
    /**
     * Alias for log, to make developer's life easier <3
     */
    get log() {
        return this.logger;
    }
    /**
     * Returns promise with the current user (if authToken argument was given)
     *
     * This is a temporary solution that decodes the userID from the access token (JWT)
     * and fetches the user info
     *
     * @returns {Promise<User>}
     */
    get $authUser() {
        const [, payload] = this.args.authToken.split('.');
        const buf = new Buffer(payload, 'base64');
        const { sub: userID } = JSON.parse(buf.toString('ascii'));
        return this.SystemAPI.userRead({ userID }).then(r => new User(r));
    }
    /**
     * Configures and returns system API client
     */
    get SystemAPI() {
        var _a, _b;
        if (!this.systemAPI) {
            if (!((_b = (_a = this.config) === null || _a === void 0 ? void 0 : _a.cServers) === null || _b === void 0 ? void 0 : _b.system)) {
                throw new Error('configuration for corteza system server missing');
            }
            this.systemAPI = new System({
                baseURL: this.config.cServers.system.apiBaseURL,
                accessTokenFn: () => this.args.authToken,
            });
        }
        return this.systemAPI;
    }
    /**
     * Configures and returns compose API client
     */
    get ComposeAPI() {
        var _a, _b;
        if (!this.composeAPI) {
            if (!((_b = (_a = this.config) === null || _a === void 0 ? void 0 : _a.cServers) === null || _b === void 0 ? void 0 : _b.compose)) {
                throw new Error('configuration for corteza compose server missing');
            }
            this.composeAPI = new Compose({
                baseURL: this.config.cServers.compose.apiBaseURL,
                accessTokenFn: () => this.args.authToken,
            });
        }
        return this.composeAPI;
    }
    /**
     * Configures and returns system helper
     */
    get System() {
        return new SystemHelper(Object.assign({ SystemAPI: this.SystemAPI }, this.args));
    }
    /**
     * Configures and returns compose helper
     */
    get Compose() {
        return new ComposeHelper(Object.assign({ ComposeAPI: this.ComposeAPI }, this.args));
    }
    /**
     *
     */
    get frontendBaseURL() {
        var _a, _b;
        return (_b = (_a = this.config) === null || _a === void 0 ? void 0 : _a.frontend) === null || _b === void 0 ? void 0 : _b.baseURL;
    }
}

var index$1 = /*#__PURE__*/Object.freeze({
  __proto__: null,
  Args: Args,
  ArgsProxy: ArgsProxy,
  ComposeHelper: ComposeHelper,
  CortezaTypes: CortezaTypes,
  Ctx: Ctx,
  Exec: Exec,
  SystemHelper: SystemHelper
});

class Workflow {
    constructor(w) {
        this.workflowID = NoID;
        this.handle = '';
        this.enabled = false;
        this.trace = false;
        this.keepSessions = 0;
        this.labels = {};
        this.meta = {
            name: '',
            description: '',
            visual: {},
            subWorkflow: false,
        };
        this.scope = undefined;
        this.steps = undefined;
        this.paths = undefined;
        this.issues = undefined;
        this.runAs = NoID;
        this.ownedBy = NoID;
        this.createdBy = NoID;
        this.updatedBy = NoID;
        this.deletedBy = NoID;
        this.createdAt = undefined;
        this.updatedAt = undefined;
        this.deletedAt = undefined;
        this.canGrant = false;
        this.canUpdateWorkflow = false;
        this.canDeleteWorkflow = false;
        this.canExecuteWorkflow = false;
        this.apply(w);
    }
    apply(w) {
        Apply(this, w, CortezaID, 'workflowID');
        Apply(this, w, String, 'handle');
        Apply(this, w, Boolean, 'enabled', 'trace');
        Apply(this, w, Number, 'keepSessions');
        Apply(this, w, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt');
        Apply(this, w, CortezaID, 'runAs', 'ownedBy', 'createdBy', 'updatedBy', 'deletedBy');
        Apply(this, w, Boolean, 'canGrant', 'canUpdateWorkflow', 'canDeleteWorkflow', 'canExecuteWorkflow');
        if (IsOf(w, 'meta')) {
            this.meta = Object.assign({}, w.meta);
        }
        if (IsOf(w, 'labels')) {
            this.labels = Object.assign({}, w.labels);
        }
        if (IsOf(w, 'scope')) {
            this.scope = w.scope;
        }
        if (IsOf(w, 'steps')) {
            this.steps = w.steps;
        }
        if (IsOf(w, 'paths')) {
            this.paths = w.paths;
        }
        if (IsOf(w, 'issues')) {
            this.issues = w.issues;
        }
    }
    /**
     * Returns resource ID
     */
    get resourceID() {
        return `${this.resourceType}:${this.workflowID}`;
    }
    /**
     * Resource type
     */
    get resourceType() {
        return 'automation:workflow';
    }
}

class Param {
    constructor(u) {
        this.name = '';
        this.types = [];
        this.required = false;
        this.isArray = false;
        this.meta = {};
        this.apply(u);
    }
    apply(u) {
        Apply(this, u, String, 'name');
        Apply(this, u, Boolean, 'required', 'isArray');
        if (u === null || u === void 0 ? void 0 : u.types) {
            this.types = u.types;
        }
        if (u === null || u === void 0 ? void 0 : u.meta) {
            this.meta = Object.assign({}, u.meta);
        }
    }
}

let Function$1 = class Function {
    constructor(u) {
        this.ref = '';
        this.kind = '';
        this.meta = {};
        this.parameters = [];
        this.results = [];
        this.labels = {};
        this.apply(u);
    }
    apply(u) {
        Apply(this, u, String, 'ref', 'kind');
        if (u === null || u === void 0 ? void 0 : u.parameters) {
            this.parameters = u.parameters.map(p => new Param(p));
        }
        if (u === null || u === void 0 ? void 0 : u.results) {
            this.results = u.results.map(p => new Param(p));
        }
        if (u === null || u === void 0 ? void 0 : u.meta) {
            this.meta = Object.assign({}, u.meta);
        }
        if (u === null || u === void 0 ? void 0 : u.labels) {
            this.labels = Object.assign({}, u.labels);
        }
    }
};

class Prompt {
    constructor(u) {
        this.ref = '';
        this.sessionID = NoID;
        this.stateID = NoID;
        this.createdAt = undefined;
        this.payload = undefined;
        this.apply(u);
    }
    apply(u) {
        Apply(this, u, CortezaID, 'sessionID', 'stateID');
        Apply(this, u, String, 'ref');
        Apply(this, u, ISO8601Date, 'createdAt');
        if (u === null || u === void 0 ? void 0 : u.payload) {
            this.payload = u.payload;
        }
    }
}

const { capitalize } = lodash;
function IsTyped(a) {
    return typeof a &&
        a === 'object' &&
        Object.prototype.hasOwnProperty.call(a, '@type') &&
        Object.prototype.hasOwnProperty.call(a, '@value');
}
function unwrap(v) {
    return IsTyped(v) ? v['@value'] : v;
}
function cast(v) {
    return { '@value': unwrap(v), '@type': guessType(v) };
}
function guessType(v) {
    switch (typeof v) {
        case 'boolean':
            return 'Boolean';
        case 'string':
            return 'String';
        case 'number':
            return Number(v) === v && v % 1 === 0 ? 'Float' : 'Integer';
        case 'object':
            if (v.resourceType) {
                // converts foo:bar into FooBar
                return v.resourceType.split(':').map(capitalize).join('');
            }
            return 'Any';
        default:
            return 'Any';
    }
}
/**
 *
 * @param any
 * @constructor
 */
function Encode(input) {
    const output = {};
    for (const key in input) {
        output[key] = IsTyped(input[key]) ? input[key] : cast(input[key]);
    }
    return output;
}

var index = /*#__PURE__*/Object.freeze({
  __proto__: null,
  Encode: Encode,
  Function: Function$1,
  IsTyped: IsTyped,
  Param: Param,
  Prompt: Prompt,
  Workflow: Workflow
});

exports.NoID = NoID;
exports.apiClients = index$2;
exports.automation = index;
exports.compose = index$3;
exports.corredor = index$1;
exports.eventbus = index$8;
exports.fmt = index$4;
exports.reporter = index$6;
exports.shared = index$7;
exports.system = index$5;
exports.validator = validator;
//# sourceMappingURL=index.cjs.map
