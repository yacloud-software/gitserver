// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/cloud/vision/v1/product_search.proto

package vision

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	_ "google.golang.org/genproto/googleapis/api/annotations"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Parameters for a product search request.
type ProductSearchParams struct {
	// The bounding polygon around the area of interest in the image.
	// Optional. If it is not specified, system discretion will be applied.
	BoundingPoly *BoundingPoly `protobuf:"bytes,9,opt,name=bounding_poly,json=boundingPoly,proto3" json:"bounding_poly,omitempty"`
	// The resource name of a [ProductSet][google.cloud.vision.v1.ProductSet] to be searched for similar images.
	//
	// Format is:
	// `projects/PROJECT_ID/locations/LOC_ID/productSets/PRODUCT_SET_ID`.
	ProductSet string `protobuf:"bytes,6,opt,name=product_set,json=productSet,proto3" json:"product_set,omitempty"`
	// The list of product categories to search in. Currently, we only consider
	// the first category, and either "homegoods-v2", "apparel-v2", "toys-v2",
	// "packagedgoods-v1", or "general-v1" should be specified. The legacy
	// categories "homegoods", "apparel", and "toys" are still supported but will
	// be deprecated. For new products, please use "homegoods-v2", "apparel-v2",
	// or "toys-v2" for better product search accuracy. It is recommended to
	// migrate existing products to these categories as well.
	ProductCategories []string `protobuf:"bytes,7,rep,name=product_categories,json=productCategories,proto3" json:"product_categories,omitempty"`
	// The filtering expression. This can be used to restrict search results based
	// on Product labels. We currently support an AND of OR of key-value
	// expressions, where each expression within an OR must have the same key. An
	// '=' should be used to connect the key and value.
	//
	// For example, "(color = red OR color = blue) AND brand = Google" is
	// acceptable, but "(color = red OR brand = Google)" is not acceptable.
	// "color: red" is not acceptable because it uses a ':' instead of an '='.
	Filter               string   `protobuf:"bytes,8,opt,name=filter,proto3" json:"filter,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProductSearchParams) Reset()         { *m = ProductSearchParams{} }
func (m *ProductSearchParams) String() string { return proto.CompactTextString(m) }
func (*ProductSearchParams) ProtoMessage()    {}
func (*ProductSearchParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_4fdf2c80d2106c63, []int{0}
}

func (m *ProductSearchParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProductSearchParams.Unmarshal(m, b)
}
func (m *ProductSearchParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProductSearchParams.Marshal(b, m, deterministic)
}
func (m *ProductSearchParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProductSearchParams.Merge(m, src)
}
func (m *ProductSearchParams) XXX_Size() int {
	return xxx_messageInfo_ProductSearchParams.Size(m)
}
func (m *ProductSearchParams) XXX_DiscardUnknown() {
	xxx_messageInfo_ProductSearchParams.DiscardUnknown(m)
}

var xxx_messageInfo_ProductSearchParams proto.InternalMessageInfo

func (m *ProductSearchParams) GetBoundingPoly() *BoundingPoly {
	if m != nil {
		return m.BoundingPoly
	}
	return nil
}

func (m *ProductSearchParams) GetProductSet() string {
	if m != nil {
		return m.ProductSet
	}
	return ""
}

func (m *ProductSearchParams) GetProductCategories() []string {
	if m != nil {
		return m.ProductCategories
	}
	return nil
}

func (m *ProductSearchParams) GetFilter() string {
	if m != nil {
		return m.Filter
	}
	return ""
}

// Results for a product search request.
type ProductSearchResults struct {
	// Timestamp of the index which provided these results. Products added to the
	// product set and products removed from the product set after this time are
	// not reflected in the current results.
	IndexTime *timestamp.Timestamp `protobuf:"bytes,2,opt,name=index_time,json=indexTime,proto3" json:"index_time,omitempty"`
	// List of results, one for each product match.
	Results []*ProductSearchResults_Result `protobuf:"bytes,5,rep,name=results,proto3" json:"results,omitempty"`
	// List of results grouped by products detected in the query image. Each entry
	// corresponds to one bounding polygon in the query image, and contains the
	// matching products specific to that region. There may be duplicate product
	// matches in the union of all the per-product results.
	ProductGroupedResults []*ProductSearchResults_GroupedResult `protobuf:"bytes,6,rep,name=product_grouped_results,json=productGroupedResults,proto3" json:"product_grouped_results,omitempty"`
	XXX_NoUnkeyedLiteral  struct{}                              `json:"-"`
	XXX_unrecognized      []byte                                `json:"-"`
	XXX_sizecache         int32                                 `json:"-"`
}

func (m *ProductSearchResults) Reset()         { *m = ProductSearchResults{} }
func (m *ProductSearchResults) String() string { return proto.CompactTextString(m) }
func (*ProductSearchResults) ProtoMessage()    {}
func (*ProductSearchResults) Descriptor() ([]byte, []int) {
	return fileDescriptor_4fdf2c80d2106c63, []int{1}
}

func (m *ProductSearchResults) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProductSearchResults.Unmarshal(m, b)
}
func (m *ProductSearchResults) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProductSearchResults.Marshal(b, m, deterministic)
}
func (m *ProductSearchResults) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProductSearchResults.Merge(m, src)
}
func (m *ProductSearchResults) XXX_Size() int {
	return xxx_messageInfo_ProductSearchResults.Size(m)
}
func (m *ProductSearchResults) XXX_DiscardUnknown() {
	xxx_messageInfo_ProductSearchResults.DiscardUnknown(m)
}

var xxx_messageInfo_ProductSearchResults proto.InternalMessageInfo

func (m *ProductSearchResults) GetIndexTime() *timestamp.Timestamp {
	if m != nil {
		return m.IndexTime
	}
	return nil
}

func (m *ProductSearchResults) GetResults() []*ProductSearchResults_Result {
	if m != nil {
		return m.Results
	}
	return nil
}

func (m *ProductSearchResults) GetProductGroupedResults() []*ProductSearchResults_GroupedResult {
	if m != nil {
		return m.ProductGroupedResults
	}
	return nil
}

// Information about a product.
type ProductSearchResults_Result struct {
	// The Product.
	Product *Product `protobuf:"bytes,1,opt,name=product,proto3" json:"product,omitempty"`
	// A confidence level on the match, ranging from 0 (no confidence) to
	// 1 (full confidence).
	Score float32 `protobuf:"fixed32,2,opt,name=score,proto3" json:"score,omitempty"`
	// The resource name of the image from the product that is the closest match
	// to the query.
	Image                string   `protobuf:"bytes,3,opt,name=image,proto3" json:"image,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProductSearchResults_Result) Reset()         { *m = ProductSearchResults_Result{} }
func (m *ProductSearchResults_Result) String() string { return proto.CompactTextString(m) }
func (*ProductSearchResults_Result) ProtoMessage()    {}
func (*ProductSearchResults_Result) Descriptor() ([]byte, []int) {
	return fileDescriptor_4fdf2c80d2106c63, []int{1, 0}
}

func (m *ProductSearchResults_Result) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProductSearchResults_Result.Unmarshal(m, b)
}
func (m *ProductSearchResults_Result) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProductSearchResults_Result.Marshal(b, m, deterministic)
}
func (m *ProductSearchResults_Result) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProductSearchResults_Result.Merge(m, src)
}
func (m *ProductSearchResults_Result) XXX_Size() int {
	return xxx_messageInfo_ProductSearchResults_Result.Size(m)
}
func (m *ProductSearchResults_Result) XXX_DiscardUnknown() {
	xxx_messageInfo_ProductSearchResults_Result.DiscardUnknown(m)
}

var xxx_messageInfo_ProductSearchResults_Result proto.InternalMessageInfo

func (m *ProductSearchResults_Result) GetProduct() *Product {
	if m != nil {
		return m.Product
	}
	return nil
}

func (m *ProductSearchResults_Result) GetScore() float32 {
	if m != nil {
		return m.Score
	}
	return 0
}

func (m *ProductSearchResults_Result) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

// Prediction for what the object in the bounding box is.
type ProductSearchResults_ObjectAnnotation struct {
	// Object ID that should align with EntityAnnotation mid.
	Mid string `protobuf:"bytes,1,opt,name=mid,proto3" json:"mid,omitempty"`
	// The BCP-47 language code, such as "en-US" or "sr-Latn". For more
	// information, see
	// http://www.unicode.org/reports/tr35/#Unicode_locale_identifier.
	LanguageCode string `protobuf:"bytes,2,opt,name=language_code,json=languageCode,proto3" json:"language_code,omitempty"`
	// Object name, expressed in its `language_code` language.
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// Score of the result. Range [0, 1].
	Score                float32  `protobuf:"fixed32,4,opt,name=score,proto3" json:"score,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProductSearchResults_ObjectAnnotation) Reset()         { *m = ProductSearchResults_ObjectAnnotation{} }
func (m *ProductSearchResults_ObjectAnnotation) String() string { return proto.CompactTextString(m) }
func (*ProductSearchResults_ObjectAnnotation) ProtoMessage()    {}
func (*ProductSearchResults_ObjectAnnotation) Descriptor() ([]byte, []int) {
	return fileDescriptor_4fdf2c80d2106c63, []int{1, 1}
}

func (m *ProductSearchResults_ObjectAnnotation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProductSearchResults_ObjectAnnotation.Unmarshal(m, b)
}
func (m *ProductSearchResults_ObjectAnnotation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProductSearchResults_ObjectAnnotation.Marshal(b, m, deterministic)
}
func (m *ProductSearchResults_ObjectAnnotation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProductSearchResults_ObjectAnnotation.Merge(m, src)
}
func (m *ProductSearchResults_ObjectAnnotation) XXX_Size() int {
	return xxx_messageInfo_ProductSearchResults_ObjectAnnotation.Size(m)
}
func (m *ProductSearchResults_ObjectAnnotation) XXX_DiscardUnknown() {
	xxx_messageInfo_ProductSearchResults_ObjectAnnotation.DiscardUnknown(m)
}

var xxx_messageInfo_ProductSearchResults_ObjectAnnotation proto.InternalMessageInfo

func (m *ProductSearchResults_ObjectAnnotation) GetMid() string {
	if m != nil {
		return m.Mid
	}
	return ""
}

func (m *ProductSearchResults_ObjectAnnotation) GetLanguageCode() string {
	if m != nil {
		return m.LanguageCode
	}
	return ""
}

func (m *ProductSearchResults_ObjectAnnotation) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ProductSearchResults_ObjectAnnotation) GetScore() float32 {
	if m != nil {
		return m.Score
	}
	return 0
}

// Information about the products similar to a single product in a query
// image.
type ProductSearchResults_GroupedResult struct {
	// The bounding polygon around the product detected in the query image.
	BoundingPoly *BoundingPoly `protobuf:"bytes,1,opt,name=bounding_poly,json=boundingPoly,proto3" json:"bounding_poly,omitempty"`
	// List of results, one for each product match.
	Results []*ProductSearchResults_Result `protobuf:"bytes,2,rep,name=results,proto3" json:"results,omitempty"`
	// List of generic predictions for the object in the bounding box.
	ObjectAnnotations    []*ProductSearchResults_ObjectAnnotation `protobuf:"bytes,3,rep,name=object_annotations,json=objectAnnotations,proto3" json:"object_annotations,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                 `json:"-"`
	XXX_unrecognized     []byte                                   `json:"-"`
	XXX_sizecache        int32                                    `json:"-"`
}

func (m *ProductSearchResults_GroupedResult) Reset()         { *m = ProductSearchResults_GroupedResult{} }
func (m *ProductSearchResults_GroupedResult) String() string { return proto.CompactTextString(m) }
func (*ProductSearchResults_GroupedResult) ProtoMessage()    {}
func (*ProductSearchResults_GroupedResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_4fdf2c80d2106c63, []int{1, 2}
}

func (m *ProductSearchResults_GroupedResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProductSearchResults_GroupedResult.Unmarshal(m, b)
}
func (m *ProductSearchResults_GroupedResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProductSearchResults_GroupedResult.Marshal(b, m, deterministic)
}
func (m *ProductSearchResults_GroupedResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProductSearchResults_GroupedResult.Merge(m, src)
}
func (m *ProductSearchResults_GroupedResult) XXX_Size() int {
	return xxx_messageInfo_ProductSearchResults_GroupedResult.Size(m)
}
func (m *ProductSearchResults_GroupedResult) XXX_DiscardUnknown() {
	xxx_messageInfo_ProductSearchResults_GroupedResult.DiscardUnknown(m)
}

var xxx_messageInfo_ProductSearchResults_GroupedResult proto.InternalMessageInfo

func (m *ProductSearchResults_GroupedResult) GetBoundingPoly() *BoundingPoly {
	if m != nil {
		return m.BoundingPoly
	}
	return nil
}

func (m *ProductSearchResults_GroupedResult) GetResults() []*ProductSearchResults_Result {
	if m != nil {
		return m.Results
	}
	return nil
}

func (m *ProductSearchResults_GroupedResult) GetObjectAnnotations() []*ProductSearchResults_ObjectAnnotation {
	if m != nil {
		return m.ObjectAnnotations
	}
	return nil
}

func init() {
	proto.RegisterType((*ProductSearchParams)(nil), "google.cloud.vision.v1.ProductSearchParams")
	proto.RegisterType((*ProductSearchResults)(nil), "google.cloud.vision.v1.ProductSearchResults")
	proto.RegisterType((*ProductSearchResults_Result)(nil), "google.cloud.vision.v1.ProductSearchResults.Result")
	proto.RegisterType((*ProductSearchResults_ObjectAnnotation)(nil), "google.cloud.vision.v1.ProductSearchResults.ObjectAnnotation")
	proto.RegisterType((*ProductSearchResults_GroupedResult)(nil), "google.cloud.vision.v1.ProductSearchResults.GroupedResult")
}

func init() {
	proto.RegisterFile("google/cloud/vision/v1/product_search.proto", fileDescriptor_4fdf2c80d2106c63)
}

var fileDescriptor_4fdf2c80d2106c63 = []byte{
	// 594 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x94, 0x4d, 0x6f, 0xd4, 0x3c,
	0x10, 0xc7, 0x95, 0xdd, 0x36, 0x7d, 0xe2, 0xb6, 0x52, 0xeb, 0xa7, 0x94, 0x10, 0x21, 0x75, 0x55,
	0xa8, 0xb4, 0x12, 0x22, 0x51, 0xdb, 0x53, 0x79, 0x39, 0x74, 0x2b, 0x51, 0x71, 0x00, 0x56, 0x06,
	0x71, 0xe0, 0x12, 0x79, 0x13, 0xd7, 0x18, 0x25, 0x99, 0xc8, 0x76, 0x56, 0x94, 0xcf, 0x81, 0xc4,
	0x9d, 0x2f, 0xc8, 0x15, 0x71, 0x42, 0xb1, 0x9d, 0xed, 0x6e, 0x69, 0x05, 0x45, 0x9c, 0x32, 0x63,
	0xff, 0xe7, 0x37, 0x9e, 0xc9, 0xd8, 0xe8, 0x01, 0x07, 0xe0, 0x05, 0x4b, 0xb2, 0x02, 0x9a, 0x3c,
	0x99, 0x0a, 0x25, 0xa0, 0x4a, 0xa6, 0xfb, 0x49, 0x2d, 0x21, 0x6f, 0x32, 0x9d, 0x2a, 0x46, 0x65,
	0xf6, 0x3e, 0xae, 0x25, 0x68, 0xc0, 0xdb, 0x56, 0x1c, 0x1b, 0x71, 0x6c, 0xc5, 0xf1, 0x74, 0x3f,
	0xda, 0xbb, 0x06, 0xc2, 0x19, 0x94, 0x4c, 0xcb, 0x73, 0x1b, 0x1e, 0x1d, 0xfe, 0x51, 0xae, 0x54,
	0x31, 0x39, 0x15, 0x19, 0x73, 0x41, 0x3b, 0x2e, 0xc8, 0x78, 0x93, 0xe6, 0x2c, 0xd1, 0xa2, 0x64,
	0x4a, 0xd3, 0xb2, 0x76, 0x82, 0xbb, 0x4e, 0x40, 0x6b, 0x91, 0xd0, 0xaa, 0x02, 0x4d, 0xb5, 0x80,
	0x4a, 0xb9, 0xdd, 0x3b, 0x73, 0xbb, 0x92, 0x29, 0x68, 0x64, 0x47, 0xde, 0xfd, 0xe6, 0xa1, 0xff,
	0xc7, 0x36, 0xf5, 0x6b, 0x93, 0x79, 0x4c, 0x25, 0x2d, 0x15, 0x7e, 0x8e, 0xd6, 0x27, 0xd0, 0x54,
	0xb9, 0xa8, 0x78, 0x5a, 0x43, 0x71, 0x1e, 0x06, 0x03, 0x6f, 0xb8, 0x7a, 0x70, 0x3f, 0xbe, 0xba,
	0xfa, 0x78, 0xe4, 0xc4, 0x63, 0x28, 0xce, 0xc9, 0xda, 0x64, 0xce, 0xc3, 0xcf, 0xd0, 0xea, 0x45,
	0x71, 0x3a, 0xf4, 0x07, 0xde, 0x30, 0x18, 0xed, 0xfd, 0x38, 0xde, 0x45, 0x03, 0x17, 0x6f, 0x99,
	0xb4, 0x16, 0x2a, 0xce, 0xa0, 0x4c, 0x66, 0xe7, 0xd1, 0x04, 0xd5, 0x33, 0x1b, 0x3f, 0x44, 0xb8,
	0xe3, 0x64, 0x54, 0x33, 0x0e, 0x52, 0x30, 0x15, 0xae, 0x0c, 0xfa, 0xc3, 0x80, 0x6c, 0xba, 0x9d,
	0x93, 0xd9, 0x06, 0xde, 0x46, 0xfe, 0x99, 0x28, 0x34, 0x93, 0xe1, 0x7f, 0x6d, 0x46, 0xe2, 0xbc,
	0xdd, 0xcf, 0x3e, 0xda, 0x5a, 0xa8, 0x98, 0x30, 0xd5, 0x14, 0x5a, 0xe1, 0x23, 0x84, 0x44, 0x95,
	0xb3, 0x8f, 0x69, 0xdb, 0xdc, 0xb0, 0x67, 0xea, 0x8d, 0xba, 0x7a, 0xbb, 0xce, 0xc7, 0x6f, 0xba,
	0xce, 0x93, 0xc0, 0xa8, 0x5b, 0x1f, 0xbf, 0x40, 0x2b, 0xd2, 0x52, 0xc2, 0xe5, 0x41, 0x7f, 0xb8,
	0x7a, 0x70, 0x78, 0x5d, 0x9f, 0xae, 0xca, 0x1c, 0xdb, 0x2f, 0xe9, 0x18, 0x58, 0xa2, 0xdb, 0x5d,
	0xa5, 0x5c, 0x42, 0x53, 0xb3, 0x3c, 0xed, 0xf0, 0xbe, 0xc1, 0x3f, 0xba, 0x11, 0xfe, 0xd4, 0x32,
	0x5c, 0x96, 0x5b, 0x0e, 0xbd, 0xb0, 0xaa, 0x22, 0x40, 0xbe, 0x35, 0xf1, 0x11, 0x5a, 0x71, 0x92,
	0xd0, 0x33, 0x4d, 0xd8, 0xf9, 0x4d, 0x36, 0xd2, 0xe9, 0xf1, 0x16, 0x5a, 0x56, 0x19, 0x48, 0xdb,
	0xbd, 0x1e, 0xb1, 0x4e, 0xbb, 0x2a, 0x4a, 0xca, 0x59, 0xd8, 0x37, 0x3f, 0xc2, 0x3a, 0x91, 0x42,
	0x1b, 0xaf, 0x26, 0x1f, 0x58, 0xa6, 0x8f, 0x67, 0xf3, 0x8a, 0x37, 0x50, 0xbf, 0x14, 0xb9, 0x49,
	0x1b, 0x90, 0xd6, 0xc4, 0xf7, 0xd0, 0x7a, 0x41, 0x2b, 0xde, 0x50, 0xce, 0xd2, 0x0c, 0x72, 0x4b,
	0x0e, 0xc8, 0x5a, 0xb7, 0x78, 0x02, 0x39, 0xc3, 0x18, 0x2d, 0x55, 0xb4, 0xec, 0xf8, 0xc6, 0xbe,
	0x38, 0xca, 0xd2, 0xdc, 0x51, 0xa2, 0x2f, 0x3d, 0xb4, 0xbe, 0x50, 0xf8, 0xaf, 0x83, 0xee, 0xfd,
	0xf5, 0xa0, 0xcf, 0x4d, 0x41, 0xef, 0x1f, 0x4c, 0x41, 0x81, 0x30, 0x98, 0x06, 0xa5, 0x73, 0x37,
	0x3a, 0xec, 0x1b, 0xf2, 0xd3, 0x1b, 0x91, 0x2f, 0xf7, 0x99, 0x6c, 0xc2, 0xa5, 0x15, 0x35, 0xfa,
	0x84, 0xa2, 0x0c, 0xca, 0x6b, 0xb0, 0x23, 0xbc, 0xf8, 0x46, 0xb4, 0x97, 0x61, 0xec, 0xbd, 0x7b,
	0xe2, 0xd4, 0x1c, 0xda, 0xdf, 0x11, 0x83, 0xe4, 0x09, 0x67, 0x95, 0xb9, 0x2a, 0xc9, 0xc5, 0x9d,
	0xbe, 0xfc, 0xd4, 0x3d, 0xb6, 0xd6, 0x77, 0xcf, 0xfb, 0xda, 0x5b, 0x3a, 0x3d, 0x79, 0xfb, 0x72,
	0xe2, 0x9b, 0x90, 0xc3, 0x9f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x17, 0x28, 0x07, 0x76, 0x88, 0x05,
	0x00, 0x00,
}
