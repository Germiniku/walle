/**
* @Time: 2021/1/28 下午1:38
* @Author: miku
* @File: protocol
* @Version: 1.0.0
* @Description:
 */

package protocol

import "fmt"

type ProducerMessage struct {
	Key   string
	Topic string
	Value []byte
}

func PackagePublishKey(topic, key string) string {
	return fmt.Sprintf("walle.%s.%s", topic, key)
}
