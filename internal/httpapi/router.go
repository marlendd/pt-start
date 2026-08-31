package httpapi

import "net/http"

func NewRouter(handler *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/shorten",
		handler.requireMethod(
			http.MethodPost,
			handler.Shorten,
		),
	)

	mux.HandleFunc(
		"/{code}",
		handler.requireMethod(
			http.MethodGet,
			handler.Redirect,
		),
	)

	mux.HandleFunc("/", handler.NotFound)

	return mux
}

func (h *Handler) requireMethod(
	method string,
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)

			h.writeError(
				r.Context(),
				w,
				http.StatusMethodNotAllowed,
				errCodeMethodNotAllowed,
				"method not allowed",
			)
			return
		}

		next(w, r)
	}
}

func (h *Handler) NotFound(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.writeError(
		r.Context(),
		w,
		http.StatusNotFound,
		errCodeNotFound,
		"resource not found",
	)
}
