package sql

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
)

// If managedMembers is non-empty: returns only members that are both in the role and in managedMembers.
// If managedMembers is nil or empty: returns all members in the role.
func (c *Connector) GetServerRoleMember(ctx context.Context, roleName string, managedMembers []string) (*model.ServerRoleMember, error) {
	cmd := `SELECT COALESCE(sp.name, sql_logins.name) AS name
			FROM [sys].[server_role_members] srm
			INNER JOIN [sys].[server_principals] role ON srm.role_principal_id = role.principal_id AND role.type = 'R' AND role.name = @roleName
			LEFT JOIN [sys].[server_principals] sp ON srm.member_principal_id = sp.principal_id
			LEFT JOIN [sys].[sql_logins] sql_logins ON srm.member_principal_id = sql_logins.principal_id
			WHERE COALESCE(sp.name, sql_logins.name) IS NOT NULL`
	var inRole []string
	err := c.
		QueryContext(ctx, cmd,
			func(r *sql.Rows) error {
				for r.Next() {
					var name string
					if err := r.Scan(&name); err != nil {
						return err
					}
					inRole = append(inRole, name)
				}
				return nil
			},
			sql.Named("roleName", roleName),
		)
	if err != nil {
		return nil, err
	}

	if len(managedMembers) == 0 {
		return &model.ServerRoleMember{RoleName: roleName, Members: inRole}, nil
	}
	managedSet := make(map[string]struct{}, len(managedMembers))
	for _, m := range managedMembers {
		managedSet[m] = struct{}{}
	}
	var members []string
	for _, name := range inRole {
		if _, ok := managedSet[name]; ok {
			members = append(members, name)
		}
	}
	return &model.ServerRoleMember{RoleName: roleName, Members: members}, nil
}

func (c *Connector) CreateServerRoleMember(ctx context.Context, roleName string, members []string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE member_cur CURSOR FOR SELECT value FROM String_Split(@members, ',')
			DECLARE @member_name nvarchar(max)
			OPEN member_cur
			FETCH NEXT FROM member_cur INTO @member_name
			WHILE @@FETCH_STATUS = 0
				BEGIN
					SET @stmt = 'ALTER SERVER ROLE ' + @roleName + ' ADD MEMBER ' + @member_name
					EXEC (@stmt)
					FETCH NEXT FROM member_cur INTO @member_name
				END
			CLOSE member_cur
			DEALLOCATE member_cur
			`
	return c.
		ExecContext(ctx, cmd,
			sql.Named("roleName", roleName),
			sql.Named("members", strings.Join(members, ",")),
		)
}

func (c *Connector) UpdateServerRoleMember(ctx context.Context, roleName string, members []string, changeType string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE member_cur CURSOR FOR SELECT value FROM String_Split(@members, ',')
			DECLARE @member_name nvarchar(max)
			OPEN member_cur
			FETCH NEXT FROM member_cur INTO @member_name
			WHILE @@FETCH_STATUS = 0
				BEGIN
					SET @stmt = 'ALTER SERVER ROLE ' + @roleName + ' ' + @changeType + ' MEMBER ' + @member_name
					EXEC (@stmt)
					FETCH NEXT FROM member_cur INTO @member_name
				END
			CLOSE member_cur
			DEALLOCATE member_cur
			`
	return c.
		ExecContext(ctx, cmd,
			sql.Named("roleName", roleName),
			sql.Named("members", strings.Join(members, ",")),
			sql.Named("changeType", changeType),
		)
}

func (c *Connector) DeleteServerRoleMember(ctx context.Context, roleName string, members []string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE member_cur CURSOR FOR SELECT value FROM String_Split(@members, ',')
			DECLARE @member_name nvarchar(max)
			OPEN member_cur
			FETCH NEXT FROM member_cur INTO @member_name
			WHILE @@FETCH_STATUS = 0
				BEGIN
					SET @stmt = 'ALTER SERVER ROLE ' + @roleName + ' DROP MEMBER ' + @member_name
					EXEC (@stmt)
					FETCH NEXT FROM member_cur INTO @member_name
				END
			CLOSE member_cur
			DEALLOCATE member_cur
			`
	return c.
		ExecContext(ctx, cmd,
			sql.Named("roleName", roleName),
			sql.Named("members", strings.Join(members, ",")),
		)
}
