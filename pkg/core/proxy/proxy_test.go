package proxy

import "testing"

func TestTemplateTypeConstants(t *testing.T) {
	templates := map[TemplateType]string{
		TemplateStatic:       "static",
		TemplateReverseProxy: "reverse_proxy",
		TemplatePHPFastCGI:   "php_fastcgi",
	}

	for template, expected := range templates {
		if string(template) != expected {
			t.Errorf("TemplateType %q != %q", template, expected)
		}
	}
}