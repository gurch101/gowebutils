package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/mailutils"
)

type InviteUserController struct {
	app *app.App
}

func NewInviteUserController(app *app.App) *InviteUserController {
	return &InviteUserController{app: app}
}

type InviteUserRequest struct {
	UserName string `json:"userName"`
	Email    string `json:"email"`
}

func (c *InviteUserController) InviteUser(w http.ResponseWriter, r *http.Request) {
	inviteUserRequest, err := httputils.ReadJSON[InviteUserRequest](w, r)
	if err != nil {
		httputils.BadRequestResponse(w, r, err)

		return
	}

	user := authutils.ContextGetUser(r)

	err = InviteUser(
		r.Context(),
		c.app.Mailer,
		c.app.GetEnvVarString("HOST"),
		user.TenantID,
		inviteUserRequest.UserName,
		inviteUserRequest.Email,
	)

	if err != nil {
		httputils.ServerErrorResponse(w, r, err)

		return
	}

	err = httputils.WriteJSON(w, http.StatusOK, nil, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func InviteUser(
	_ context.Context,
	mailer mailutils.Mailer,
	hostName string,
	tenantID int64,
	username, email string,
) error {
	payload := map[string]any{
		"tenant_id": tenantID,
		"email":     email,
		"username":  username,
	}
	inviteToken, err := authutils.CreateInviteToken(payload)

	if err != nil {
		return err
	}

	mailer.Send(email, "invite.go.tmpl", map[string]string{
		"URL": fmt.Sprintf("%s/register?code=%s", hostName, inviteToken),
	})

	return nil
}
