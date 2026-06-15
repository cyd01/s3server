package auth

import "net/http"

func buildCanonicalRequest(
	r *http.Request,
	signedHeaders []string,
) (string, error) {
	// méthode
	// URI
	// query triée
	// headers normalisés
	// signed headers
	// payload hash

	/*TODO*/
	canonical := ""

	return canonical, nil
}
