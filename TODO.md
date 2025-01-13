# TODO

- [ ] Play with the gpg stuff: <https://github.com/openpgpjs/openpgpjs>
- [ ] Client tool (which should be callable from the docker as well) for admin user management
- [ ] Auth on the API and UI
- [ ] Theme management from: <https://claude.ai/chat/e560622b-357a-47db-900b-2e1fa29dc25b> to avoid below

```
TypeError: bgColor is null index.js:943:21
    hasBuiltInDarkTheme moz-extension://201bc7f3-197e-4e3c-b383-516aa8105d5e/inject/index.js:943
    runCheck moz-extension://201bc7f3-197e-4e3c-b383-516aa8105d5e/inject/index.js:1004
    1 moz-extension://201bc7f3-197e-4e3c-b383-516aa8105d5e/inject/index.js:1079
    (Async: MutationCallback)
    runDarkThemeDetector moz-extension://201bc7f3-197e-4e3c-b383-516aa8105d5e/inject/index.js:1076
    onMessage moz-extension://201bc7f3-197e-4e3c-b383-516aa8105d5e/inject/index.js:8765
    apply self-hosted:2285
    raw resource://gre/modules/ExtensionCommon.sys.mjs:2847
    wrapResponse resource://gre/modules/ExtensionChild.sys.mjs:207
    responses resource://gre/modules/ExtensionChild.sys.mjs:176
    map self-hosted:175
    emit resource://gre/modules/ExtensionChild.sys.mjs:176
    recvRuntimeMessage resource://gre/modules/ExtensionChild.sys.mjs:409
    _recv resource://gre/modules/ConduitsChild.sys.mjs:90
    receiveMessage resource://gre/modules/ConduitsChild.sys.mjs:201

```

- [ ] Establish proper navigation and footer from snakey-py
- [ ] Combine SPA feel with templ somehow - since nesting is fast AF in templ
- [ ] Start porting the admin interface
- [ ] Use websocket shit instead of socketio for group chats
- [ ] Create a redis new database (with a seperate connection string, for scaling out later) for the TTL messages
