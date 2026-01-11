package mme_agent

func CallOnLoad(agent IBaseAgent, new bool) error {
	if lifecycle, ok := agent.(interface{ OnLoad(new bool) error }); ok {
		return lifecycle.OnLoad(new)
	}
	return nil
}

func CallOnSave(agent IBaseAgent) error {
	if lifecycle, ok := agent.(interface{ OnSave() error }); ok {
		return lifecycle.OnSave()
	}
	return nil
}

func CallOnLogin(agent IBaseAgent) error {
	if lifecycle, ok := agent.(interface{ OnLogin() error }); ok {
		return lifecycle.OnLogin()
	}
	return nil
}

func CallOnLogout(agent IBaseAgent) error {
	if lifecycle, ok := agent.(interface{ OnLogout() error }); ok {
		return lifecycle.OnLogout()
	}
	return nil
}
