package controller

import "github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"

// ProcessAudit process the given audit and save it
// TODO: implement the function
func (c *Controller) ProcessAudit(ar *models.AuditRecord) {
	c.log.DebugStruct("the following record should be processed", ar)
}
