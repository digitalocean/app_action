package parser_struct

//for dockerhub integration updatedRepo will look like
// type UpdatedRepo struct {
// 	Registry_type string
// 	Name       string
// 	Registry   string
// 	Repository string
// 	Tag        string
// }
// UpdatedRepo used for parsing json object of changed repo
type UpdatedRepo struct {
	Name       string
	Repository string
	Tag        string
}
