package query

import (
	"context"
	"database/sql"
	errs "errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/zitadel/logging"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/api/call"
	"github.com/zitadel/zitadel/internal/crypto"
	"github.com/zitadel/zitadel/internal/database"
	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/query/projection"
	"github.com/zitadel/zitadel/internal/repository/idp"
	"github.com/zitadel/zitadel/internal/telemetry/tracing"
)

type IDPTemplate struct {
	CreationDate      time.Time
	ChangeDate        time.Time
	Sequence          uint64
	ResourceOwner     string
	ID                string
	State             domain.IDPState
	Name              string
	Type              domain.IDPType
	OwnerType         domain.IdentityProviderType
	IsCreationAllowed bool
	IsLinkingAllowed  bool
	IsAutoCreation    bool
	IsAutoUpdate      bool
	*OAuthIDPTemplate
	*OIDCIDPTemplate
	*JWTIDPTemplate
	*GoogleIDPTemplate
	*LDAPIDPTemplate
}

type IDPTemplates struct {
	SearchResponse
	Templates []*IDPTemplate
}

type OAuthIDPTemplate struct {
	IDPID                 string
	ClientID              string
	ClientSecret          *crypto.CryptoValue
	AuthorizationEndpoint string
	TokenEndpoint         string
	UserEndpoint          string
	Scopes                database.StringArray
	IDAttribute           string
}

type OIDCIDPTemplate struct {
	IDPID        string
	ClientID     string
	ClientSecret *crypto.CryptoValue
	Issuer       string
	Scopes       database.StringArray
}

type JWTIDPTemplate struct {
	IDPID        string
	Issuer       string
	KeysEndpoint string
	HeaderName   string
	Endpoint     string
}

type GoogleIDPTemplate struct {
	IDPID        string
	ClientID     string
	ClientSecret *crypto.CryptoValue
	Scopes       database.StringArray
}

type LDAPIDPTemplate struct {
	IDPID               string
	Host                string
	Port                string
	TLS                 bool
	BaseDN              string
	UserObjectClass     string
	UserUniqueAttribute string
	Admin               string
	Password            *crypto.CryptoValue
	idp.LDAPAttributes
}

var (
	idpTemplateTable = table{
		name:          projection.IDPTemplateTable,
		instanceIDCol: projection.IDPTemplateInstanceIDCol,
	}
	IDPTemplateIDCol = Column{
		name:  projection.IDPTemplateIDCol,
		table: idpTemplateTable,
	}
	IDPTemplateCreationDateCol = Column{
		name:  projection.IDPTemplateCreationDateCol,
		table: idpTemplateTable,
	}
	IDPTemplateChangeDateCol = Column{
		name:  projection.IDPTemplateChangeDateCol,
		table: idpTemplateTable,
	}
	IDPTemplateSequenceCol = Column{
		name:  projection.IDPTemplateSequenceCol,
		table: idpTemplateTable,
	}
	IDPTemplateResourceOwnerCol = Column{
		name:  projection.IDPTemplateResourceOwnerCol,
		table: idpTemplateTable,
	}
	IDPTemplateInstanceIDCol = Column{
		name:  projection.IDPTemplateInstanceIDCol,
		table: idpTemplateTable,
	}
	IDPTemplateStateCol = Column{
		name:  projection.IDPTemplateStateCol,
		table: idpTemplateTable,
	}
	IDPTemplateNameCol = Column{
		name:  projection.IDPTemplateNameCol,
		table: idpTemplateTable,
	}
	IDPTemplateOwnerTypeCol = Column{
		name:  projection.IDPOwnerTypeCol,
		table: idpTemplateTable,
	}
	IDPTemplateTypeCol = Column{
		name:  projection.IDPTemplateTypeCol,
		table: idpTemplateTable,
	}
	IDPTemplateOwnerRemovedCol = Column{
		name:  projection.IDPTemplateOwnerRemovedCol,
		table: idpTemplateTable,
	}
	IDPTemplateIsCreationAllowedCol = Column{
		name:  projection.IDPTemplateIsCreationAllowedCol,
		table: idpTemplateTable,
	}
	IDPTemplateIsLinkingAllowedCol = Column{
		name:  projection.IDPTemplateIsLinkingAllowedCol,
		table: idpTemplateTable,
	}
	IDPTemplateIsAutoCreationCol = Column{
		name:  projection.IDPTemplateIsAutoCreationCol,
		table: idpTemplateTable,
	}
	IDPTemplateIsAutoUpdateCol = Column{
		name:  projection.IDPTemplateIsAutoUpdateCol,
		table: idpTemplateTable,
	}
)

var (
	oauthIdpTemplateTable = table{
		name:          projection.IDPTemplateOAuthTable,
		instanceIDCol: projection.OAuthInstanceIDCol,
	}
	OAuthIDCol = Column{
		name:  projection.OAuthIDCol,
		table: oauthIdpTemplateTable,
	}
	OAuthInstanceIDCol = Column{
		name:  projection.OAuthInstanceIDCol,
		table: oauthIdpTemplateTable,
	}
	OAuthClientIDCol = Column{
		name:  projection.OAuthClientIDCol,
		table: oauthIdpTemplateTable,
	}
	OAuthClientSecretCol = Column{
		name:  projection.OAuthClientSecretCol,
		table: oauthIdpTemplateTable,
	}
	OAuthAuthorizationEndpointCol = Column{
		name:  projection.OAuthAuthorizationEndpointCol,
		table: oauthIdpTemplateTable,
	}
	OAuthTokenEndpointCol = Column{
		name:  projection.OAuthTokenEndpointCol,
		table: oauthIdpTemplateTable,
	}
	OAuthUserEndpointCol = Column{
		name:  projection.OAuthUserEndpointCol,
		table: oauthIdpTemplateTable,
	}
	OAuthScopesCol = Column{
		name:  projection.OAuthScopesCol,
		table: oauthIdpTemplateTable,
	}
	OAuthIDAttributeCol = Column{
		name:  projection.OAuthIDAttributeCol,
		table: oauthIdpTemplateTable,
	}
)

var (
	oidcIdpTemplateTable = table{
		name:          projection.IDPTemplateOIDCTable,
		instanceIDCol: projection.OIDCInstanceIDCol,
	}
	OIDCIDCol = Column{
		name:  projection.OIDCIDCol,
		table: oidcIdpTemplateTable,
	}
	OIDCInstanceIDCol = Column{
		name:  projection.OIDCInstanceIDCol,
		table: oidcIdpTemplateTable,
	}
	OIDCIssuerCol = Column{
		name:  projection.OIDCIssuerCol,
		table: oidcIdpTemplateTable,
	}
	OIDCClientIDCol = Column{
		name:  projection.OIDCClientIDCol,
		table: oidcIdpTemplateTable,
	}
	OIDCClientSecretCol = Column{
		name:  projection.OIDCClientSecretCol,
		table: oidcIdpTemplateTable,
	}
	OIDCScopesCol = Column{
		name:  projection.OIDCScopesCol,
		table: oidcIdpTemplateTable,
	}
)

var (
	jwtIdpTemplateTable = table{
		name:          projection.IDPTemplateJWTTable,
		instanceIDCol: projection.JWTInstanceIDCol,
	}
	JWTIDCol = Column{
		name:  projection.JWTIDCol,
		table: jwtIdpTemplateTable,
	}
	JWTInstanceIDCol = Column{
		name:  projection.JWTInstanceIDCol,
		table: jwtIdpTemplateTable,
	}
	JWTIssuerCol = Column{
		name:  projection.JWTIssuerCol,
		table: jwtIdpTemplateTable,
	}
	JWTEndpointCol = Column{
		name:  projection.JWTEndpointCol,
		table: jwtIdpTemplateTable,
	}
	JWTKeysEndpointCol = Column{
		name:  projection.JWTKeysEndpointCol,
		table: jwtIdpTemplateTable,
	}
	JWTHeaderNameCol = Column{
		name:  projection.JWTHeaderNameCol,
		table: jwtIdpTemplateTable,
	}
)

var (
	googleIdpTemplateTable = table{
		name:          projection.IDPTemplateGoogleTable,
		instanceIDCol: projection.GoogleInstanceIDCol,
	}
	GoogleIDCol = Column{
		name:  projection.GoogleIDCol,
		table: googleIdpTemplateTable,
	}
	GoogleInstanceIDCol = Column{
		name:  projection.GoogleInstanceIDCol,
		table: googleIdpTemplateTable,
	}
	GoogleClientIDCol = Column{
		name:  projection.GoogleClientIDCol,
		table: googleIdpTemplateTable,
	}
	GoogleClientSecretCol = Column{
		name:  projection.GoogleClientSecretCol,
		table: googleIdpTemplateTable,
	}
	GoogleScopesCol = Column{
		name:  projection.GoogleScopesCol,
		table: googleIdpTemplateTable,
	}
)

var (
	ldapIdpTemplateTable = table{
		name:          projection.IDPTemplateLDAPTable,
		instanceIDCol: projection.IDPTemplateInstanceIDCol,
	}
	LDAPIDCol = Column{
		name:  projection.LDAPIDCol,
		table: ldapIdpTemplateTable,
	}
	LDAPInstanceIDCol = Column{
		name:  projection.LDAPInstanceIDCol,
		table: ldapIdpTemplateTable,
	}
	LDAPHostCol = Column{
		name:  projection.LDAPHostCol,
		table: ldapIdpTemplateTable,
	}
	LDAPPortCol = Column{
		name:  projection.LDAPPortCol,
		table: ldapIdpTemplateTable,
	}
	LDAPTlsCol = Column{
		name:  projection.LDAPTlsCol,
		table: ldapIdpTemplateTable,
	}
	LDAPBaseDNCol = Column{
		name:  projection.LDAPBaseDNCol,
		table: ldapIdpTemplateTable,
	}
	LDAPUserObjectClassCol = Column{
		name:  projection.LDAPUserObjectClassCol,
		table: ldapIdpTemplateTable,
	}
	LDAPUserUniqueAttributeCol = Column{
		name:  projection.LDAPUserUniqueAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPAdminCol = Column{
		name:  projection.LDAPAdminCol,
		table: ldapIdpTemplateTable,
	}
	LDAPPasswordCol = Column{
		name:  projection.LDAPPasswordCol,
		table: ldapIdpTemplateTable,
	}
	LDAPIDAttributeCol = Column{
		name:  projection.LDAPIDAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPFirstNameAttributeCol = Column{
		name:  projection.LDAPFirstNameAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPLastNameAttributeCol = Column{
		name:  projection.LDAPLastNameAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPDisplayNameAttributeCol = Column{
		name:  projection.LDAPDisplayNameAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPNickNameAttributeCol = Column{
		name:  projection.LDAPNickNameAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPPreferredUsernameAttributeCol = Column{
		name:  projection.LDAPPreferredUsernameAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPEmailAttributeCol = Column{
		name:  projection.LDAPEmailAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPEmailVerifiedAttributeCol = Column{
		name:  projection.LDAPEmailVerifiedAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPPhoneAttributeCol = Column{
		name:  projection.LDAPPhoneAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPPhoneVerifiedAttributeCol = Column{
		name:  projection.LDAPPhoneVerifiedAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPPreferredLanguageAttributeCol = Column{
		name:  projection.LDAPPreferredLanguageAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPAvatarURLAttributeCol = Column{
		name:  projection.LDAPAvatarURLAttributeCol,
		table: ldapIdpTemplateTable,
	}
	LDAPProfileAttributeCol = Column{
		name:  projection.LDAPProfileAttributeCol,
		table: ldapIdpTemplateTable,
	}
)

// IDPTemplateByID searches for the requested id
func (q *Queries) IDPTemplateByID(ctx context.Context, shouldTriggerBulk bool, id string, withOwnerRemoved bool, queries ...SearchQuery) (_ *IDPTemplate, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	if shouldTriggerBulk {
		err := projection.IDPTemplateProjection.Trigger(ctx)
		logging.OnError(err).WithField("projection", idpTemplateTable.identifier()).Warn("could not trigger projection for query")
	}

	eq := sq.Eq{
		IDPTemplateIDCol.identifier():         id,
		IDPTemplateInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
	}
	if !withOwnerRemoved {
		eq[IDPTemplateOwnerRemovedCol.identifier()] = false
	}
	query, scan := prepareIDPTemplateByIDQuery(ctx, q.client)
	for _, q := range queries {
		query = q.toQuery(query)
	}
	stmt, args, err := query.Where(eq).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-SFefg", "Errors.Query.SQLStatement")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

// IDPTemplates searches idp templates matching the query
func (q *Queries) IDPTemplates(ctx context.Context, queries *IDPTemplateSearchQueries, withOwnerRemoved bool) (idps *IDPTemplates, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	query, scan := prepareIDPTemplatesQuery(ctx, q.client)
	eq := sq.Eq{
		IDPTemplateInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
	}
	if !withOwnerRemoved {
		eq[IDPTemplateOwnerRemovedCol.identifier()] = false
	}
	stmt, args, err := queries.toQuery(query).Where(eq).ToSql()
	if err != nil {
		return nil, errors.ThrowInvalidArgument(err, "QUERY-SAF34", "Errors.Query.InvalidRequest")
	}

	rows, err := q.client.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-BDFrq", "Errors.Internal")
	}
	idps, err = scan(rows)
	if err != nil {
		return nil, err
	}
	idps.LatestSequence, err = q.latestSequence(ctx, idpTemplateTable)
	return idps, err
}

type IDPTemplateSearchQueries struct {
	SearchRequest
	Queries []SearchQuery
}

func NewIDPTemplateIDSearchQuery(id string) (SearchQuery, error) {
	return NewTextQuery(IDPTemplateIDCol, id, TextEquals)
}

func NewIDPTemplateOwnerTypeSearchQuery(ownerType domain.IdentityProviderType) (SearchQuery, error) {
	return NewNumberQuery(IDPTemplateOwnerTypeCol, ownerType, NumberEquals)
}

func NewIDPTemplateNameSearchQuery(method TextComparison, value string) (SearchQuery, error) {
	return NewTextQuery(IDPTemplateNameCol, value, method)
}

func NewIDPTemplateResourceOwnerSearchQuery(value string) (SearchQuery, error) {
	return NewTextQuery(IDPTemplateResourceOwnerCol, value, TextEquals)
}

func NewIDPTemplateResourceOwnerListSearchQuery(ids ...string) (SearchQuery, error) {
	list := make([]interface{}, len(ids))
	for i, value := range ids {
		list[i] = value
	}
	return NewListQuery(IDPTemplateResourceOwnerCol, list, ListIn)
}

func (q *IDPTemplateSearchQueries) toQuery(query sq.SelectBuilder) sq.SelectBuilder {
	query = q.SearchRequest.toQuery(query)
	for _, q := range q.Queries {
		query = q.toQuery(query)
	}
	return query
}

func prepareIDPTemplateByIDQuery(ctx context.Context, db prepareDatabase) (sq.SelectBuilder, func(*sql.Row) (*IDPTemplate, error)) {
	return sq.Select(
			IDPTemplateIDCol.identifier(),
			IDPTemplateResourceOwnerCol.identifier(),
			IDPTemplateCreationDateCol.identifier(),
			IDPTemplateChangeDateCol.identifier(),
			IDPTemplateSequenceCol.identifier(),
			IDPTemplateStateCol.identifier(),
			IDPTemplateNameCol.identifier(),
			IDPTemplateTypeCol.identifier(),
			IDPTemplateOwnerTypeCol.identifier(),
			IDPTemplateIsCreationAllowedCol.identifier(),
			IDPTemplateIsLinkingAllowedCol.identifier(),
			IDPTemplateIsAutoCreationCol.identifier(),
			IDPTemplateIsAutoUpdateCol.identifier(),
			// oauth
			OAuthIDCol.identifier(),
			OAuthClientIDCol.identifier(),
			OAuthClientSecretCol.identifier(),
			OAuthAuthorizationEndpointCol.identifier(),
			OAuthTokenEndpointCol.identifier(),
			OAuthUserEndpointCol.identifier(),
			OAuthScopesCol.identifier(),
			OAuthIDAttributeCol.identifier(),
			// oidc
			OIDCIDCol.identifier(),
			OIDCIssuerCol.identifier(),
			OIDCClientIDCol.identifier(),
			OIDCClientSecretCol.identifier(),
			OIDCScopesCol.identifier(),
			// jwt
			JWTIDCol.identifier(),
			JWTIssuerCol.identifier(),
			JWTEndpointCol.identifier(),
			JWTKeysEndpointCol.identifier(),
			JWTHeaderNameCol.identifier(),
			// google
			GoogleIDCol.identifier(),
			GoogleClientIDCol.identifier(),
			GoogleClientSecretCol.identifier(),
			GoogleScopesCol.identifier(),
			// ldap
			LDAPIDCol.identifier(),
			LDAPHostCol.identifier(),
			LDAPPortCol.identifier(),
			LDAPTlsCol.identifier(),
			LDAPBaseDNCol.identifier(),
			LDAPUserObjectClassCol.identifier(),
			LDAPUserUniqueAttributeCol.identifier(),
			LDAPAdminCol.identifier(),
			LDAPPasswordCol.identifier(),
			LDAPIDAttributeCol.identifier(),
			LDAPFirstNameAttributeCol.identifier(),
			LDAPLastNameAttributeCol.identifier(),
			LDAPDisplayNameAttributeCol.identifier(),
			LDAPNickNameAttributeCol.identifier(),
			LDAPPreferredUsernameAttributeCol.identifier(),
			LDAPEmailAttributeCol.identifier(),
			LDAPEmailVerifiedAttributeCol.identifier(),
			LDAPPhoneAttributeCol.identifier(),
			LDAPPhoneVerifiedAttributeCol.identifier(),
			LDAPPreferredLanguageAttributeCol.identifier(),
			LDAPAvatarURLAttributeCol.identifier(),
			LDAPProfileAttributeCol.identifier(),
		).From(idpTemplateTable.identifier()).
			LeftJoin(join(OAuthIDCol, IDPTemplateIDCol)).
			LeftJoin(join(OIDCIDCol, IDPTemplateIDCol)).
			LeftJoin(join(JWTIDCol, IDPTemplateIDCol)).
			LeftJoin(join(GoogleIDCol, IDPTemplateIDCol)).
			LeftJoin(join(LDAPIDCol, IDPTemplateIDCol) + db.Timetravel(call.Took(ctx))).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*IDPTemplate, error) {
			idpTemplate := new(IDPTemplate)

			name := sql.NullString{}

			oauthID := sql.NullString{}
			oauthClientID := sql.NullString{}
			oauthClientSecret := new(crypto.CryptoValue)
			oauthAuthorizationEndpoint := sql.NullString{}
			oauthTokenEndpoint := sql.NullString{}
			oauthUserEndpoint := sql.NullString{}
			oauthScopes := database.StringArray{}
			oauthIDAttribute := sql.NullString{}

			oidcID := sql.NullString{}
			oidcIssuer := sql.NullString{}
			oidcClientID := sql.NullString{}
			oidcClientSecret := new(crypto.CryptoValue)
			oidcScopes := database.StringArray{}

			jwtID := sql.NullString{}
			jwtIssuer := sql.NullString{}
			jwtEndpoint := sql.NullString{}
			jwtKeysEndpoint := sql.NullString{}
			jwtHeaderName := sql.NullString{}

			googleID := sql.NullString{}
			googleClientID := sql.NullString{}
			googleClientSecret := new(crypto.CryptoValue)
			googleScopes := database.StringArray{}

			ldapID := sql.NullString{}
			ldapHost := sql.NullString{}
			ldapPort := sql.NullString{}
			ldapTls := sql.NullBool{}
			ldapBaseDN := sql.NullString{}
			ldapUserObjectClass := sql.NullString{}
			ldapUserUniqueAttribute := sql.NullString{}
			ldapAdmin := sql.NullString{}
			ldapPassword := new(crypto.CryptoValue)
			ldapIDAttribute := sql.NullString{}
			ldapFirstNameAttribute := sql.NullString{}
			ldapLastNameAttribute := sql.NullString{}
			ldapDisplayNameAttribute := sql.NullString{}
			ldapNickNameAttribute := sql.NullString{}
			ldapPreferredUsernameAttribute := sql.NullString{}
			ldapEmailAttribute := sql.NullString{}
			ldapEmailVerifiedAttribute := sql.NullString{}
			ldapPhoneAttribute := sql.NullString{}
			ldapPhoneVerifiedAttribute := sql.NullString{}
			ldapPreferredLanguageAttribute := sql.NullString{}
			ldapAvatarURLAttribute := sql.NullString{}
			ldapProfileAttribute := sql.NullString{}

			err := row.Scan(
				&idpTemplate.ID,
				&idpTemplate.ResourceOwner,
				&idpTemplate.CreationDate,
				&idpTemplate.ChangeDate,
				&idpTemplate.Sequence,
				&idpTemplate.State,
				&name,
				&idpTemplate.Type,
				&idpTemplate.OwnerType,
				&idpTemplate.IsCreationAllowed,
				&idpTemplate.IsLinkingAllowed,
				&idpTemplate.IsAutoCreation,
				&idpTemplate.IsAutoUpdate,
				// oauth
				&oauthID,
				&oauthClientID,
				&oauthClientSecret,
				&oauthAuthorizationEndpoint,
				&oauthTokenEndpoint,
				&oauthUserEndpoint,
				&oauthScopes,
				&oauthIDAttribute,
				// oidc
				&oidcID,
				&oidcIssuer,
				&oidcClientID,
				&oidcClientSecret,
				&oidcScopes,
				// jwt
				&jwtID,
				&jwtIssuer,
				&jwtEndpoint,
				&jwtKeysEndpoint,
				&jwtHeaderName,
				// google
				&googleID,
				&googleClientID,
				&googleClientSecret,
				&googleScopes,
				// ldap
				&ldapID,
				&ldapHost,
				&ldapPort,
				&ldapTls,
				&ldapBaseDN,
				&ldapUserObjectClass,
				&ldapUserUniqueAttribute,
				&ldapAdmin,
				&ldapPassword,
				&ldapIDAttribute,
				&ldapFirstNameAttribute,
				&ldapLastNameAttribute,
				&ldapDisplayNameAttribute,
				&ldapNickNameAttribute,
				&ldapPreferredUsernameAttribute,
				&ldapEmailAttribute,
				&ldapEmailVerifiedAttribute,
				&ldapPhoneAttribute,
				&ldapPhoneVerifiedAttribute,
				&ldapPreferredLanguageAttribute,
				&ldapAvatarURLAttribute,
				&ldapProfileAttribute,
			)
			if err != nil {
				if errs.Is(err, sql.ErrNoRows) {
					return nil, errors.ThrowNotFound(err, "QUERY-SAFrt", "Errors.IDPConfig.NotExisting")
				}
				return nil, errors.ThrowInternal(err, "QUERY-ADG42", "Errors.Internal")
			}

			idpTemplate.Name = name.String

			if oauthID.Valid {
				idpTemplate.OAuthIDPTemplate = &OAuthIDPTemplate{
					IDPID:                 oauthID.String,
					ClientID:              oauthClientID.String,
					ClientSecret:          oauthClientSecret,
					AuthorizationEndpoint: oauthAuthorizationEndpoint.String,
					TokenEndpoint:         oauthTokenEndpoint.String,
					UserEndpoint:          oauthUserEndpoint.String,
					Scopes:                oauthScopes,
					IDAttribute:           oauthIDAttribute.String,
				}
			}
			if oidcID.Valid {
				idpTemplate.OIDCIDPTemplate = &OIDCIDPTemplate{
					IDPID:        oidcID.String,
					ClientID:     oidcClientID.String,
					ClientSecret: oidcClientSecret,
					Issuer:       oidcIssuer.String,
					Scopes:       oidcScopes,
				}
			}
			if jwtID.Valid {
				idpTemplate.JWTIDPTemplate = &JWTIDPTemplate{
					IDPID:        jwtID.String,
					Issuer:       jwtIssuer.String,
					KeysEndpoint: jwtKeysEndpoint.String,
					HeaderName:   jwtHeaderName.String,
					Endpoint:     jwtEndpoint.String,
				}
			}
			if googleID.Valid {
				idpTemplate.GoogleIDPTemplate = &GoogleIDPTemplate{
					IDPID:        googleID.String,
					ClientID:     googleClientID.String,
					ClientSecret: googleClientSecret,
					Scopes:       googleScopes,
				}
			}
			if ldapID.Valid {
				idpTemplate.LDAPIDPTemplate = &LDAPIDPTemplate{
					IDPID:               ldapID.String,
					Host:                ldapHost.String,
					Port:                ldapPort.String,
					TLS:                 ldapTls.Bool,
					BaseDN:              ldapBaseDN.String,
					UserObjectClass:     ldapUserObjectClass.String,
					UserUniqueAttribute: ldapUserUniqueAttribute.String,
					Admin:               ldapAdmin.String,
					Password:            ldapPassword,
					LDAPAttributes: idp.LDAPAttributes{
						IDAttribute:                ldapIDAttribute.String,
						FirstNameAttribute:         ldapFirstNameAttribute.String,
						LastNameAttribute:          ldapLastNameAttribute.String,
						DisplayNameAttribute:       ldapDisplayNameAttribute.String,
						NickNameAttribute:          ldapNickNameAttribute.String,
						PreferredUsernameAttribute: ldapPreferredUsernameAttribute.String,
						EmailAttribute:             ldapEmailAttribute.String,
						EmailVerifiedAttribute:     ldapEmailVerifiedAttribute.String,
						PhoneAttribute:             ldapPhoneAttribute.String,
						PhoneVerifiedAttribute:     ldapPhoneVerifiedAttribute.String,
						PreferredLanguageAttribute: ldapPreferredLanguageAttribute.String,
						AvatarURLAttribute:         ldapAvatarURLAttribute.String,
						ProfileAttribute:           ldapProfileAttribute.String,
					},
				}
			}

			return idpTemplate, nil
		}
}

func prepareIDPTemplatesQuery(ctx context.Context, db prepareDatabase) (sq.SelectBuilder, func(*sql.Rows) (*IDPTemplates, error)) {
	return sq.Select(
			IDPTemplateIDCol.identifier(),
			IDPTemplateResourceOwnerCol.identifier(),
			IDPTemplateCreationDateCol.identifier(),
			IDPTemplateChangeDateCol.identifier(),
			IDPTemplateSequenceCol.identifier(),
			IDPTemplateStateCol.identifier(),
			IDPTemplateNameCol.identifier(),
			IDPTemplateTypeCol.identifier(),
			IDPTemplateOwnerTypeCol.identifier(),
			IDPTemplateIsCreationAllowedCol.identifier(),
			IDPTemplateIsLinkingAllowedCol.identifier(),
			IDPTemplateIsAutoCreationCol.identifier(),
			IDPTemplateIsAutoUpdateCol.identifier(),
			// oauth
			OAuthIDCol.identifier(),
			OAuthClientIDCol.identifier(),
			OAuthClientSecretCol.identifier(),
			OAuthAuthorizationEndpointCol.identifier(),
			OAuthTokenEndpointCol.identifier(),
			OAuthUserEndpointCol.identifier(),
			OAuthScopesCol.identifier(),
			OAuthIDAttributeCol.identifier(),
			// oidc
			OIDCIDCol.identifier(),
			OIDCIssuerCol.identifier(),
			OIDCClientIDCol.identifier(),
			OIDCClientSecretCol.identifier(),
			OIDCScopesCol.identifier(),
			// jwt
			JWTIDCol.identifier(),
			JWTIssuerCol.identifier(),
			JWTEndpointCol.identifier(),
			JWTKeysEndpointCol.identifier(),
			JWTHeaderNameCol.identifier(),
			// google
			GoogleIDCol.identifier(),
			GoogleClientIDCol.identifier(),
			GoogleClientSecretCol.identifier(),
			GoogleScopesCol.identifier(),
			// ldap
			LDAPIDCol.identifier(),
			LDAPHostCol.identifier(),
			LDAPPortCol.identifier(),
			LDAPTlsCol.identifier(),
			LDAPBaseDNCol.identifier(),
			LDAPUserObjectClassCol.identifier(),
			LDAPUserUniqueAttributeCol.identifier(),
			LDAPAdminCol.identifier(),
			LDAPPasswordCol.identifier(),
			LDAPIDAttributeCol.identifier(),
			LDAPFirstNameAttributeCol.identifier(),
			LDAPLastNameAttributeCol.identifier(),
			LDAPDisplayNameAttributeCol.identifier(),
			LDAPNickNameAttributeCol.identifier(),
			LDAPPreferredUsernameAttributeCol.identifier(),
			LDAPEmailAttributeCol.identifier(),
			LDAPEmailVerifiedAttributeCol.identifier(),
			LDAPPhoneAttributeCol.identifier(),
			LDAPPhoneVerifiedAttributeCol.identifier(),
			LDAPPreferredLanguageAttributeCol.identifier(),
			LDAPAvatarURLAttributeCol.identifier(),
			LDAPProfileAttributeCol.identifier(),
			countColumn.identifier(),
		).From(idpTemplateTable.identifier()).
			LeftJoin(join(OAuthIDCol, IDPTemplateIDCol)).
			LeftJoin(join(OIDCIDCol, IDPTemplateIDCol)).
			LeftJoin(join(JWTIDCol, IDPTemplateIDCol)).
			LeftJoin(join(GoogleIDCol, IDPTemplateIDCol)).
			LeftJoin(join(LDAPIDCol, IDPTemplateIDCol) + db.Timetravel(call.Took(ctx))).
			PlaceholderFormat(sq.Dollar),
		func(rows *sql.Rows) (*IDPTemplates, error) {
			templates := make([]*IDPTemplate, 0)
			var count uint64
			for rows.Next() {
				idpTemplate := new(IDPTemplate)

				name := sql.NullString{}

				oauthID := sql.NullString{}
				oauthClientID := sql.NullString{}
				oauthClientSecret := new(crypto.CryptoValue)
				oauthAuthorizationEndpoint := sql.NullString{}
				oauthTokenEndpoint := sql.NullString{}
				oauthUserEndpoint := sql.NullString{}
				oauthScopes := database.StringArray{}
				oauthIDAttribute := sql.NullString{}

				oidcID := sql.NullString{}
				oidcIssuer := sql.NullString{}
				oidcClientID := sql.NullString{}
				oidcClientSecret := new(crypto.CryptoValue)
				oidcScopes := database.StringArray{}

				jwtID := sql.NullString{}
				jwtIssuer := sql.NullString{}
				jwtEndpoint := sql.NullString{}
				jwtKeysEndpoint := sql.NullString{}
				jwtHeaderName := sql.NullString{}

				googleID := sql.NullString{}
				googleClientID := sql.NullString{}
				googleClientSecret := new(crypto.CryptoValue)
				googleScopes := database.StringArray{}

				ldapID := sql.NullString{}
				ldapHost := sql.NullString{}
				ldapPort := sql.NullString{}
				ldapTls := sql.NullBool{}
				ldapBaseDN := sql.NullString{}
				ldapUserObjectClass := sql.NullString{}
				ldapUserUniqueAttribute := sql.NullString{}
				ldapAdmin := sql.NullString{}
				ldapPassword := new(crypto.CryptoValue)
				ldapIDAttribute := sql.NullString{}
				ldapFirstNameAttribute := sql.NullString{}
				ldapLastNameAttribute := sql.NullString{}
				ldapDisplayNameAttribute := sql.NullString{}
				ldapNickNameAttribute := sql.NullString{}
				ldapPreferredUsernameAttribute := sql.NullString{}
				ldapEmailAttribute := sql.NullString{}
				ldapEmailVerifiedAttribute := sql.NullString{}
				ldapPhoneAttribute := sql.NullString{}
				ldapPhoneVerifiedAttribute := sql.NullString{}
				ldapPreferredLanguageAttribute := sql.NullString{}
				ldapAvatarURLAttribute := sql.NullString{}
				ldapProfileAttribute := sql.NullString{}

				err := rows.Scan(
					&idpTemplate.ID,
					&idpTemplate.ResourceOwner,
					&idpTemplate.CreationDate,
					&idpTemplate.ChangeDate,
					&idpTemplate.Sequence,
					&idpTemplate.State,
					&name,
					&idpTemplate.Type,
					&idpTemplate.OwnerType,
					&idpTemplate.IsCreationAllowed,
					&idpTemplate.IsLinkingAllowed,
					&idpTemplate.IsAutoCreation,
					&idpTemplate.IsAutoUpdate,
					// oauth
					&oauthID,
					&oauthClientID,
					&oauthClientSecret,
					&oauthAuthorizationEndpoint,
					&oauthTokenEndpoint,
					&oauthUserEndpoint,
					&oauthScopes,
					&oauthIDAttribute,
					// oidc
					&oidcID,
					&oidcIssuer,
					&oidcClientID,
					&oidcClientSecret,
					&oidcScopes,
					// jwt
					&jwtID,
					&jwtIssuer,
					&jwtEndpoint,
					&jwtKeysEndpoint,
					&jwtHeaderName,
					// google
					&googleID,
					&googleClientID,
					&googleClientSecret,
					&googleScopes,
					// ldap
					&ldapID,
					&ldapHost,
					&ldapPort,
					&ldapTls,
					&ldapBaseDN,
					&ldapUserObjectClass,
					&ldapUserUniqueAttribute,
					&ldapAdmin,
					&ldapPassword,
					&ldapIDAttribute,
					&ldapFirstNameAttribute,
					&ldapLastNameAttribute,
					&ldapDisplayNameAttribute,
					&ldapNickNameAttribute,
					&ldapPreferredUsernameAttribute,
					&ldapEmailAttribute,
					&ldapEmailVerifiedAttribute,
					&ldapPhoneAttribute,
					&ldapPhoneVerifiedAttribute,
					&ldapPreferredLanguageAttribute,
					&ldapAvatarURLAttribute,
					&ldapProfileAttribute,
					&count,
				)

				if err != nil {
					return nil, err
				}

				idpTemplate.Name = name.String

				if oauthID.Valid {
					idpTemplate.OAuthIDPTemplate = &OAuthIDPTemplate{
						IDPID:                 oauthID.String,
						ClientID:              oauthClientID.String,
						ClientSecret:          oauthClientSecret,
						AuthorizationEndpoint: oauthAuthorizationEndpoint.String,
						TokenEndpoint:         oauthTokenEndpoint.String,
						UserEndpoint:          oauthUserEndpoint.String,
						Scopes:                oauthScopes,
						IDAttribute:           oauthIDAttribute.String,
					}
				}
				if oidcID.Valid {
					idpTemplate.OIDCIDPTemplate = &OIDCIDPTemplate{
						IDPID:        oidcID.String,
						ClientID:     oidcClientID.String,
						ClientSecret: oidcClientSecret,
						Issuer:       oidcIssuer.String,
						Scopes:       oidcScopes,
					}
				}
				if jwtID.Valid {
					idpTemplate.JWTIDPTemplate = &JWTIDPTemplate{
						IDPID:        jwtID.String,
						Issuer:       jwtIssuer.String,
						KeysEndpoint: jwtKeysEndpoint.String,
						HeaderName:   jwtHeaderName.String,
						Endpoint:     jwtEndpoint.String,
					}
				}
				if googleID.Valid {
					idpTemplate.GoogleIDPTemplate = &GoogleIDPTemplate{
						IDPID:        googleID.String,
						ClientID:     googleClientID.String,
						ClientSecret: googleClientSecret,
						Scopes:       googleScopes,
					}
				}
				if ldapID.Valid {
					idpTemplate.LDAPIDPTemplate = &LDAPIDPTemplate{
						IDPID:               ldapID.String,
						Host:                ldapHost.String,
						Port:                ldapPort.String,
						TLS:                 ldapTls.Bool,
						BaseDN:              ldapBaseDN.String,
						UserObjectClass:     ldapUserObjectClass.String,
						UserUniqueAttribute: ldapUserUniqueAttribute.String,
						Admin:               ldapAdmin.String,
						Password:            ldapPassword,
						LDAPAttributes: idp.LDAPAttributes{
							IDAttribute:                ldapIDAttribute.String,
							FirstNameAttribute:         ldapFirstNameAttribute.String,
							LastNameAttribute:          ldapLastNameAttribute.String,
							DisplayNameAttribute:       ldapDisplayNameAttribute.String,
							NickNameAttribute:          ldapNickNameAttribute.String,
							PreferredUsernameAttribute: ldapPreferredUsernameAttribute.String,
							EmailAttribute:             ldapEmailAttribute.String,
							EmailVerifiedAttribute:     ldapEmailVerifiedAttribute.String,
							PhoneAttribute:             ldapPhoneAttribute.String,
							PhoneVerifiedAttribute:     ldapPhoneVerifiedAttribute.String,
							PreferredLanguageAttribute: ldapPreferredLanguageAttribute.String,
							AvatarURLAttribute:         ldapAvatarURLAttribute.String,
							ProfileAttribute:           ldapProfileAttribute.String,
						},
					}
				}
				templates = append(templates, idpTemplate)
			}

			if err := rows.Close(); err != nil {
				return nil, errors.ThrowInternal(err, "QUERY-SAGrt", "Errors.Query.CloseRows")
			}

			return &IDPTemplates{
				Templates: templates,
				SearchResponse: SearchResponse{
					Count: count,
				},
			}, nil
		}
}