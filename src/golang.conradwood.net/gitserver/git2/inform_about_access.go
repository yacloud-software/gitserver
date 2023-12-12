package git2

import (
	"fmt"
	slack "golang.conradwood.net/apis/slackgateway"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
)

func (h *HTTPRequest) informAdminsAboutCommit() {
	u := h.user
	s := "<no user>"
	if u != nil {
		s = fmt.Sprintf("User %s [%s]", u.Email, u.ID)
		if u.ID == "1" || u.ID == "7" || u.ID == "3539" {
			fmt.Printf("No admin message, because userid == %s\n", u.ID)
			return
		}
	}
	r := fmt.Sprintf("%d/%s@%s", h.repo.gitrepo.ID, h.repo.gitrepo.ArtefactName, h.r.URL.Path)
	text := "New git commit by " + s + " to repository " + r
	p := &slack.PostRequest{UserID: "7", Text: text}
	go func(pr *slack.PostRequest) {
		ctx := authremote.Context()
		_, err := slack.GetSlackGatewayClient().Post(ctx, pr)
		if err != nil {
			fmt.Printf("Failed to slack \"%s\" to \"%s\": %s", pr.Text, pr.UserID, utils.ErrorString(err))
		}
	}(p)

}



