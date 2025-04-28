package xerr

import (
	"strings"
	"sync"
)

var i18nCache sync.Map

func i18nLookup(key string, lang string) string {
	if lang == "" {
		lang = DefaultLang
	}

	cacheKey := key + ":" + lang
	if val, ok := i18nCache.Load(cacheKey); ok {
		return val.(string)
	}

	registryMu.RLock()
	def, ok := Registry[key]
	registryMu.RUnlock()
	if !ok {
		return ""
	}

	if val, ok := def.I18n[lang]; ok {
		i18nCache.Store(cacheKey, val)
		return val
	}

	// Support for "en-US" -> fallback to "en"
	if dash := strings.Index(lang, "-"); dash != -1 {
		base := lang[:dash]
		if val, ok := def.I18n[base]; ok {
			i18nCache.Store(cacheKey, val)
			return val
		}
	}

	i18nCache.Store(cacheKey, def.Default)
	return def.Default
}
