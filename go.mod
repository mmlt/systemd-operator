module github.com/mmlt/systemd-operator

require (
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973
	github.com/davecgh/go-spew v0.0.0-20170626231645-782f4967f2dc
	github.com/ghodss/yaml v0.0.0-20150909031657-73d445a93680
	github.com/gogo/protobuf v0.0.0-20170330071051-c0656edd0d9e
	github.com/golang/glog v0.0.0-20141105023935-44145f04b68c
	github.com/golang/groupcache v0.0.0-20160516000752-02826c3e7903
	github.com/golang/protobuf v0.0.0-20171021043952-1643683e1b54
	github.com/google/gofuzz v0.0.0-20161122191042-44d81051d367
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d
	github.com/hashicorp/golang-lru v0.0.0-20160207214719-a0d98a5f2880
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/imdario/mergo v0.0.0-20141206190957-6633656539c1
	github.com/json-iterator/go v0.0.0-20171212105241-13f86432b882
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mmlt/sshclient v0.0.0
	github.com/prometheus/client_golang v0.0.0-20180623155954-77e8f2ddcfed
	github.com/prometheus/client_model v0.0.0-20171117100541-99fa1f4be8e5
	github.com/prometheus/common v0.0.0-20180518154759-7600349dcfe1
	github.com/prometheus/procfs v0.0.0-20180612222113-7d6f385de8be
	github.com/scottdware/go-bigip v0.0.0-20180518145131-fb6f8eaea132
	github.com/spf13/pflag v0.0.0-20171106142849-4c012f6dcd95
	golang.org/x/crypto v0.0.0-20170825220121-81e90905daef
	golang.org/x/net v0.0.0-20170809000501-1c05540f6879
	golang.org/x/sys v0.0.0-20171031081856-95c657629925
	golang.org/x/text v0.0.0-20170810154203-b19bf474d317
	golang.org/x/time v0.0.0-20161028155119-f51c12702a4d
	gopkg.in/inf.v0 v0.9.0
	gopkg.in/yaml.v2 v2.0.0-20170721113624-670d4cfef054
	k8s.io/api v0.0.0-20180308224125-73d903622b73
	k8s.io/apimachinery v0.0.0-20180228050457-302974c03f7e
	k8s.io/client-go v7.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20180216212618-50ae88d24ede
)

replace github.com/mmlt/sshclient v0.0.0 => ../sshclient
