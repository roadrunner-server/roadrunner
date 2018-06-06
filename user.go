package roadrunner

import (
	"os/user"
	"fmt"
)

// resolveUser attempt to find system user by it's name or uid.
func resolveUser(u string) (usr *user.User, err error) {
	usr, err = user.LookupId(u)
	if usr != nil {
		return usr, nil
	}

	return user.Lookup(u)
}

// resolveUser attempt to find system group by it's name or uid.
func resolveGroup(g string) (grp *user.Group, err error) {
	grp, err = user.LookupGroupId(g)
	if grp != nil && grp.Name != "nogroup" {
		return grp, nil
	}

	grp, err = user.LookupGroup(g)
	if grp != nil && grp.Name != "nogroup" {
		return grp, nil
	}

	return nil, fmt.Errorf("no such group %s", g)
}
