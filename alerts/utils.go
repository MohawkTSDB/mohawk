package alerts

func (alert *Alert) checkOutOfRange(value float64) bool {
	var res bool
	switch alert.r.t {
	case BETWEEN:
		if (value < alert.r.from || value > alert.r.to) {
			res = true
		} else {
			res = false
		}
	case LOWER_THAN:
		if (value > alert.r.to) {
			res = true
		} else {
			res =  false
		}
	case HIGHER_THAN:
		if (value < alert.r.from) {
			res = true
		} else {
			res = false
		}
	}
	return res
}

func (alert *Alert) setStatus(status bool){
	alert.state = status
}

func (alert *Alert) getAlertId() string {
	return alert.id
}

func (Alerts Alerts) Open(alertList []Alert) {}


