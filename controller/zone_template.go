package controller

const ServiceZoneTemplate = `
{{.origin}} {{.ttl}} IN SOA ns.dns.{{.origin}} hostmaster.{{.origin}} 1981616 1800 900 604800 86400
{{.origin}} {{.ttl}} IN NS ns.dns.{{.origin}}
ns.dns.{{.origin}} {{.ttl}} IN A {{.clusterDnsService}}
dns-version.{{.origin}} {{.ttl}} IN TXT {{.dnsSchemaVersion}}
`
const ServiceReverseZoneTemplate = `
{{.origin}} {{.ttl}} IN SOA ns.dns.{{.origin}} hostmaster.{{.origin}} 1981616 1800 900 604800 86400
{{.origin}} {{.ttl}} IN NS ns.dns.{{.origin}}
ns.dns.{{.origin}} {{.ttl}} IN A {{.clusterDnsService}}
`
const PodReverseZoneTemplate = `
{{.origin}} {{.ttl}} IN SOA ns.dns.{{.origin}} hostmaster.{{.origin}} 1981616 1800 900 604800 86400
{{.origin}} {{.ttl}} IN NS ns.dns.{{.origin}}
ns.dns.{{.origin}} {{.ttl}} IN A {{.clusterDnsService}}
`
