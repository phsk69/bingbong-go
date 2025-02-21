# TODO

- [ ] Clear any errors logged from UI
- [ ] Get deployment working via forgejo (build/push/deploy docker) to the app server
  - [ ] Use systemd env file
  - [ ] Do we compose?
  - [ ] nginx config
  - [ ] HAproxy config
- [ ] Start porting the admin pages, BUT - see if we can combine links and htmx to have fast SSR but, without going all in on SPA (ie. links to the pages as anchors that work, and dont re-render the entire fucking dom)
