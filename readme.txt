This service serves git via http.
why?

* It authenticates using "auth" service
* checks permissions to repositories using objectauth service
* keeps a database of commits
* triggers builds for each commit


BUILDRULES:

TARGETS=linux,darwin
will build linux & darwing















