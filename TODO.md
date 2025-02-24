# TODO

<<<<<<< HEAD
- [ ] Logout thing somewhere
- [ ] Generate and host API spec automagically
- [ ] Footer: Licenses, link to api spec
- [ ] Properly render admin status on the user admin page
- [ ] promotion action for users
- [ ] Group admin routes
- [ ] User profile page/routes
  - [ ] User should be able to generate and store API keys to be used by the sidecar thing
- [ ] Interactions
  - [ ] Users should be able to send a token down to the sidecar, to be stored automatically
  - [ ] Users should be able to start an encrypted chat to one or more asc keys through the UI
  - [ ] Users should be able to request the public key of another user
  - [ ] Users should be able to invite to groups etc for the funny haha meme chat
  - [ ] Users should be able to invite other named users for a 1v1 chat or an arena
  - [ ] Users should be able to invite an entire group to a private chat
=======
- [ ] Clear any errors logged from UI
- [ ] Get deployment working via forgejo (build/push/deploy docker) to the app server
  - [ ] Use systemd env file
  - [ ] Do we compose?
  - [ ] nginx config
  - [ ] HAproxy config
- [ ] Start porting the admin pages, BUT - see if we can combine links and htmx to have fast SSR but, without going all in on SPA (ie. links to the pages as anchors that work, and dont re-render the entire fucking dom)
>>>>>>> 01b0de951b60ec13e6f02921abbb313c23c46201
