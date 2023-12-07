package git2

import (
	"context"
	"fmt"
	oa "golang.conradwood.net/apis/objectauth"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/utils"
)

func HasServiceReadAccess(ctx context.Context, ot oa.OBJECTTYPE) bool {
	ar := get_service_access_rights(ctx, ot)
	if ar == nil {
		return false
	}
	return ar.ReadAccess
}
func HasServiceWriteAccess(ctx context.Context, ot oa.OBJECTTYPE) bool {
	ar := get_service_access_rights(ctx, ot)
	if ar == nil {
		return false
	}
	return ar.WriteAccess
}
func HasServiceAnyAccess(ctx context.Context, ot oa.OBJECTTYPE) bool {
	ar := get_service_access_rights(ctx, ot)
	if ar == nil {
		return false
	}
	return ar.ReadAccess || ar.WriteAccess
}
func get_service_access_rights(ctx context.Context, ot oa.OBJECTTYPE) *oa.AllAccessResponse {
	svc := auth.GetService(ctx)
	if svc == nil {
		return nil
	}

	req := &oa.AllAccessRequest{
		ObjectType: ot,
		ServiceID:  svc.ID,
	}
	res, err := oa.GetObjectAuthClient().AllowAllServiceAccess(ctx, req)
	if err != nil {
		fmt.Printf("failed to get service access right: %s\n", utils.ErrorString(err))
		return nil
	}
	//fmt.Printf("ServiceAccess: %#v\n", res)
	return res
}

