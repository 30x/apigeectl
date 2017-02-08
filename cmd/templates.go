package cmd

var TARGET_ENDPOINT = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<TargetEndpoint name="default">
    <Description/>
    <FaultRules/>
    <PreFlow name="PreFlow">
        <Request/>
        <Response>
            <Step>
                <Name>AddCors</Name>
            </Step>
        </Response>
    </PreFlow>
    <PostFlow name="PostFlow">
        <Request/>
        <Response/>
    </PostFlow>
    <Flows/>
    <ScriptContainer>
        <ResourceURL>{{.TargetPath}}</ResourceURL>
    </ScriptContainer>
</TargetEndpoint>`

var PROXY_ENDPOINT = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<ProxyEndpoint name="default">
    <Description/>
    <FaultRules/>
    <PreFlow name="PreFlow">
        <Request/>
        <Response/>
    </PreFlow>
    <PostFlow name="PostFlow">
        <Request/>
        <Response/>
    </PostFlow>
    <Flows/>
    <HTTPProxyConnection>
        <BasePath>{{.BasePath}}</BasePath>
        <Properties/>
        <VirtualHost>default</VirtualHost>
        <VirtualHost>secure</VirtualHost>
    </HTTPProxyConnection>
    <RouteRule name="default">
        <TargetEndpoint>default</TargetEndpoint>
    </RouteRule>
</ProxyEndpoint>`

var ADD_CORS = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<AssignMessage async="false" continueOnError="false" enabled="true" name="AddCors">
    <DisplayName>Add CORS Headers</DisplayName>
    <FaultRules/>
    <Properties/>
    <Add>
        <Headers>
            <Header name="Access-Control-Allow-Origin">*</Header>
            <Header name="Access-Control-Allow-Headers">origin, x-requested-with, accept</Header>
            <Header name="Access-Control-Max-Age">3628800</Header>
            <Header name="Access-Control-Allow-Methods">GET, PUT, POST, DELETE</Header>
        </Headers>
    </Add>
    <IgnoreUnresolvedVariables>true</IgnoreUnresolvedVariables>
    <AssignTo createNew="false" transport="http" type="response"/>
</AssignMessage>`

var PROXY_XML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<APIProxy revision="1" name="{{.Name}}">
    <ConfigurationVersion majorVersion="4" minorVersion="0"/>
    <CreatedAt>1459886430613</CreatedAt>
    <CreatedBy>shipyard@apigee.com</CreatedBy>
    <Description>This is a proxy for {{.Name}}, deployed on Shipyard.</Description>
    <DisplayName>{{.Name}}</DisplayName>
    <LastModifiedAt>1459886430613</LastModifiedAt>
    <LastModifiedBy>shipyard@apigee.com</LastModifiedBy>
    <Policies>
        <Policy>AddCors</Policy>
    </Policies>
    <ProxyEndpoints>
        <ProxyEndpoint>default</ProxyEndpoint>
    </ProxyEndpoints>
    <Resources/>
    <TargetServers/>
    <TargetEndpoints>
        <TargetEndpoint>default</TargetEndpoint>
    </TargetEndpoints>
    <validate>false</validate>
</APIProxy>
`

var GET_APP = `AVAILABLE REVISIONS
{{ range .}}{{.revision}}
{{end}}`

var GET_APPS = `AVAILABLE APPLICATIONS
{{ range .}}{{.name}}
{{end}}`

var GET_APP_REV = `REVISION | CREATED | IMAGE-ID
{{.revision}} | {{.created}} | {{.imageId}}`

var GET_DEPS = `NAME | CREATED | APP-REVISION
{{ range .items }}{{.metadata.name}} | {{.metadata.creationTimestamp}} | {{revision .metadata.labels}}
{{end}}`

var GET_DEP = `NAME | CREATED | APP-REVISION
{{.metadata.name}} | {{.metadata.creationTimestamp}} | {{revision .metadata.labels}}`
