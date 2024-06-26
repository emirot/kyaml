// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"sort"

	"github.com/emirot/kyaml/fieldmeta"
	"github.com/emirot/kyaml/openapi"
	"github.com/emirot/kyaml/sets"
	"github.com/emirot/kyaml/yaml"
)

// walkMap returns the value of VisitMap
//
// - call VisitMap
// - set the return value on l.Dest
// - walk each source field
// - set each source field value on l.Dest
func (l Walker) walkMap() (*yaml.RNode, error) {
	// get the new map value
	dest, err := l.Sources.setDestNode(l.VisitMap(l.Sources, l.Schema))
	if dest == nil || err != nil {
		return nil, err
	}

	// recursively set the field values on the map
	for _, key := range l.fieldNames() {
		var res *yaml.RNode
		var keys []*yaml.RNode
		if l.VisitKeysAsScalars {
			// visit the map keys as if they were scalars,
			// this is necessary if doing things such as copying
			// comments
			for i := range l.Sources {
				// construct the sources from the keys
				if l.Sources[i] == nil {
					keys = append(keys, nil)
					continue
				}
				field := l.Sources[i].Field(key)
				if field == nil || yaml.IsMissingOrNull(field.Key) {
					keys = append(keys, nil)
					continue
				}
				keys = append(keys, field.Key)
			}
			// visit the sources as a scalar
			// keys don't have any schema --pass in nil
			res, err = l.Visitor.VisitScalar(keys, nil)
			if err != nil {
				return nil, err
			}
		}

		var s *openapi.ResourceSchema
		if l.Schema != nil {
			s = l.Schema.Field(key)
		}
		fv, commentSch := l.fieldValue(key)
		if commentSch != nil {
			s = commentSch
		}
		val, err := Walker{
			VisitKeysAsScalars:    l.VisitKeysAsScalars,
			InferAssociativeLists: l.InferAssociativeLists,
			Visitor:               l,
			Schema:                s,
			Sources:               fv,
			MergeOptions:          l.MergeOptions,
			Path:                  append(l.Path, key)}.Walk()
		if err != nil {
			return nil, err
		}

		// transfer the comments of res to dest node
		var comments yaml.Comments
		if !yaml.IsMissingOrNull(res) {
			comments = yaml.Comments{
				LineComment: res.YNode().LineComment,
				HeadComment: res.YNode().HeadComment,
				FootComment: res.YNode().FootComment,
			}
			if len(keys) > 0 && !yaml.IsMissingOrNull(keys[DestIndex]) {
				keys[DestIndex].YNode().HeadComment = res.YNode().HeadComment
				keys[DestIndex].YNode().LineComment = res.YNode().LineComment
				keys[DestIndex].YNode().FootComment = res.YNode().FootComment
			}
		}

		// this handles empty and non-empty values
		_, err = dest.Pipe(yaml.FieldSetter{Name: key, Comments: comments, Value: val})
		if err != nil {
			return nil, err
		}
	}

	return dest, nil
}

// valueIfPresent returns node.Value if node is non-nil, otherwise returns nil
func (l Walker) valueIfPresent(node *yaml.MapNode) (*yaml.RNode, *openapi.ResourceSchema) {
	if node == nil {
		return nil, nil
	}

	// parse the schema for the field if present
	var s *openapi.ResourceSchema
	fm := fieldmeta.FieldMeta{}
	var err error
	// check the value for a schema
	if err = fm.Read(node.Value); err == nil {
		s = &openapi.ResourceSchema{Schema: &fm.Schema}
		if fm.Schema.Ref.String() != "" {
			r, err := openapi.Resolve(&fm.Schema.Ref, openapi.Schema())
			if err == nil && r != nil {
				s.Schema = r
			}
		}
	}
	// check the key for a schema -- this will be used
	// when the value is a Sequence (comments are attached)
	// to the key
	if fm.IsEmpty() {
		if err = fm.Read(node.Key); err == nil {
			s = &openapi.ResourceSchema{Schema: &fm.Schema}
		}
		if fm.Schema.Ref.String() != "" {
			r, err := openapi.Resolve(&fm.Schema.Ref, openapi.Schema())
			if err == nil && r != nil {
				s.Schema = r
			}
		}
	}
	return node.Value, s
}

// fieldNames returns a sorted slice containing the names of all fields that appear in any of
// the sources
func (l Walker) fieldNames() []string {
	fields := sets.String{}
	for _, s := range l.Sources {
		if s == nil {
			continue
		}
		// don't check error, we know this is a mapping node
		sFields, _ := s.Fields()
		fields.Insert(sFields...)
	}
	result := fields.List()
	sort.Strings(result)
	return result
}

// fieldValue returns a slice containing each source's value for fieldName
func (l Walker) fieldValue(fieldName string) ([]*yaml.RNode, *openapi.ResourceSchema) {
	var fields []*yaml.RNode
	var sch *openapi.ResourceSchema
	for i := range l.Sources {
		if l.Sources[i] == nil {
			fields = append(fields, nil)
			continue
		}
		field := l.Sources[i].Field(fieldName)
		f, s := l.valueIfPresent(field)
		fields = append(fields, f)
		if sch == nil && !s.IsMissingOrNull() {
			sch = s
		}
	}
	return fields, sch
}
