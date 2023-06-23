This service serves git via http.
why?

* It authenticates using "auth" service
* checks permissions to repositories using objectauth service
* keeps a database of commits
* triggers builds for each commit


BUILDRULES:

TARGETS=linux,darwin
will build linux & darwing




--- tech details --

hooks:
on access, the gitserver installs scripts in the git repository hooks directory.
these server-side scripts call into the gitserver

the post-receive uses a dodgy tcp protocol to call into the server
the update hook currently runs within the "git" process space, but an experimental implementation
to run via gRPC on the local gitserver is implemented in update.go (disabled by default)

--- a typical /etc/gitconfig looks like this --
[credential]
        helper="/usr/local/bin/gitcredentials-client -registry=registry"
        useHttpPath = true
[pull]
        rebase = false
[safe]
        directory = *



---- gitbuilder ---
the gitbuilder does builds (not gitserver).
the gitserver includes a binary called "gitcredentials".
this binary can be used as git helper. it reads context from environment variables, does a little ping-pong between gitserver and itself and passes a git compatible username/password combination to git. Git uses that to clone/push. These credentials are temporary only.



---- gitbuilder ---
the gitbuilder does builds (not gitserver).
the gitserver includes a binary called "gitcredentials".
this binary can be used as git helper. it reads context from environment variables, does a little ping-pong between gitserver and itself and passes a git compatible username/password combination to git. Git uses that to clone/push. These credentials are temporary only.







