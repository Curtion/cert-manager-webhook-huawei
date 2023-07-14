package main

import (
	"encoding/json"
	"fmt"
	"os"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}
	cmd.RunWebhookServer(GroupName,
		&huaweiDNSProviderSolver{},
	)
}

type huaweiDNSProviderSolver struct {
	client   *kubernetes.Clientset
	hwClient *dns.DnsClient
}

type huaweiDNSProviderConfig struct {
	AK     string `json:"AK"`
	SK     string `json:"SK"`
	Region string `json:"region"`
}

func (c *huaweiDNSProviderSolver) Name() string {
	return "huawei-solver"
}

// 待创建钩子, 需要允许相同值的记录多次调用
// 1. 根据ResolvedFQDN查询记录
// 2. 如果记录不存在, 创建记录
// 3. 如果记录存在, 判断记录值是否存在, 如果存在直接返回, 否则添加记录值
func (c *huaweiDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}
	c.hwClient = createHuaweiClient(cfg.AK, cfg.SK, cfg.Region)
	record, err := c.ShowRecordSet(ch.ResolvedFQDN)
	if err != nil {
		return err
	}
	if record.Id == nil {
		zeroId, err := c.GetZeroId(ch.ResolvedZone)
		if err != nil {
			return err
		}
		err = c.CreateRecordSet(zeroId, ch.ResolvedFQDN, ch.Key)
		if err != nil {
			return err
		}
	} else {
		for _, recordValue := range *record.Records {
			if recordValue == fmt.Sprintf("\"%s\"", ch.Key) {
				return nil
			}
		}
		c.UpdateRecordSet(*record.ZoneId, *record.Id, ch.ResolvedFQDN, append(*record.Records, fmt.Sprintf("\"%s\"", ch.Key)))
	}
	return nil
}

// 待删除钩子, 如果只有一条记录直接删除记录,否则只删除当前当前记录行
// 1. 根据ResolvedFQDN查询记录
// 2. 如果记录不存在, 直接返回
// 3. 如果记录存在, 则遍历记录值, 删除当前记录值再更新记录
// 4. 如果记录存在, 且记录值只有一个, 则删除记录
func (c *huaweiDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	record, err := c.ShowRecordSet(ch.ResolvedFQDN)
	if err != nil {
		return err
	}
	if record.Id == nil {
		return nil
	} else {
		if len(*record.Records) == 1 {
			c.DeleteRecordSet(*record.ZoneId, *record.Id)
		} else {
			var newRecords []string
			for _, recordValue := range *record.Records {
				if recordValue != fmt.Sprintf("\"%s\"", ch.Key) {
					newRecords = append(newRecords, recordValue)
				}
			}
			c.UpdateRecordSet(*record.ZoneId, *record.Id, ch.ResolvedFQDN, newRecords)
		}
	}
	return nil
}

// 初始化钩子
func (c *huaweiDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}
	c.client = cl
	return nil
}

// 获取域名ID
func (c *huaweiDNSProviderSolver) GetZeroId(ResolvedZone string) (string, error) {
	request := &model.ListPublicZonesRequest{}
	limitRequest := int32(1)
	request.Limit = &limitRequest
	nameRequest := ResolvedZone
	request.Name = &nameRequest
	response, err := c.hwClient.ListPublicZones(request)
	if err == nil {
		for _, zone := range *response.Zones {
			return *zone.Id, nil
		}
	}
	return "", err
}

// 创建解析记录
func (c *huaweiDNSProviderSolver) CreateRecordSet(ZoneId string, ResolvedFQDN string, Key string) error {
	request := &model.CreateRecordSetRequest{}
	request.ZoneId = ZoneId
	var listRecordsbody = []string{
		fmt.Sprintf("\"%s\"", Key),
	}
	request.Body = &model.CreateRecordSetRequestBody{
		Records: listRecordsbody,
		Type:    "TXT",
		Name:    ResolvedFQDN,
	}
	_, err := c.hwClient.CreateRecordSet(request)
	return err
}

// 更新解析记录
func (c *huaweiDNSProviderSolver) UpdateRecordSet(ZoneId string, RecordsetId string, ResolvedFQDN string, Keys []string) error {
	request := &model.UpdateRecordSetRequest{}
	request.ZoneId = ZoneId
	request.RecordsetId = RecordsetId
	var listRecordsbody = Keys
	request.Body = &model.UpdateRecordSetReq{
		Records: &listRecordsbody,
		Type:    "TXT",
		Name:    "ResolvedFQDN",
	}
	_, err := c.hwClient.UpdateRecordSet(request)
	return err
}

// 删除解析记录
func (c *huaweiDNSProviderSolver) DeleteRecordSet(ZoneId string, RecordsetId string) error {
	request := &model.DeleteRecordSetRequest{}
	request.ZoneId = ZoneId
	request.RecordsetId = RecordsetId
	_, err := c.hwClient.DeleteRecordSet(request)
	return err
}

// 获取解析记录
func (c *huaweiDNSProviderSolver) ShowRecordSet(ResolvedFQDN string) (model.ListRecordSetsWithTags, error) {
	request := &model.ListRecordSetsRequest{}
	limitRequest := int32(1)
	request.Limit = &limitRequest
	nameRequest := ResolvedFQDN
	request.Name = &nameRequest
	response, err := c.hwClient.ListRecordSets(request)
	if err == nil {
		for _, zone := range *response.Recordsets {
			return zone, nil
		}
	}
	return model.ListRecordSetsWithTags{}, err
}

// 初始化华为云客户端
func createHuaweiClient(ak string, sk string, regionName string) *dns.DnsClient {
	auth := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()
	client := dns.NewDnsClient(
		dns.DnsClientBuilder().
			WithEndpoints([]string{fmt.Sprintf("https://dns.%s.myhuaweicloud.com", regionName)}).
			WithCredential(auth).
			Build())
	return client
}

// 加载配置
func loadConfig(cfgJSON *extapi.JSON) (huaweiDNSProviderConfig, error) {
	cfg := huaweiDNSProviderConfig{}
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("solver配置解析错误: %v", err)
	}
	return cfg, nil
}
