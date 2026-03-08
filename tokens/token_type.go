package tokens

// TokenType identifies the lexical category of a token.
// Values start at 1; 0 is the invalid / zero value.
type TokenType int

const (
	// Punctuation / operators
	LParen           TokenType = iota + 1 // (
	RParen                                // )
	LBracket                              // [
	RBracket                              // ]
	LBrace                                // {
	RBrace                                // }
	Comma                                 // ,
	Dot                                   // .
	Dash                                  // -
	Plus                                  // +
	Colon                                 // :
	DotColon                              // .:
	DotCaret                              // .^
	DColon                                // ::
	DColonDollar                          // ::$
	DColonPercent                         // ::%
	DColonQMark                           // ::?
	DQMark                                // ??
	Semicolon                             // ;
	Star                                  // *
	Backslash                             // \
	Slash                                 // /
	Lt                                    // <
	Lte                                   // <=
	Gt                                    // >
	Gte                                   // >=
	Not                                   // !
	Eq                                    // =
	Neq                                   // !=  <>
	NullsafeEq                            // <=>
	ColonEq                               // :=
	ColonGt                               // :>
	NColonGt                              // !:>
	And                                   // keyword AND
	Or                                    // keyword OR
	Amp                                   // &
	DPipe                                 // ||
	PipeGt                                // |>
	Pipe                                  // |
	PipeSlash                             // |/
	DPipeSlash                            // ||/
	Caret                                 // ^
	CaretAt                               // ^@
	Tilde                                 // ~
	Arrow                                 // ->
	DArrow                                // ->>
	FArrow                                // =>
	Hash                                  // #
	HashArrow                             // #>
	DHashArrow                            // #>>
	LRArrow                               // <->
	DAT                                   // @@
	LtAt                                  // <@
	AtGt                                  // @>
	Dollar                                // $
	Parameter                             // @
	SessionToken                          // SESSION (operator context)
	SessionParameter                      // @@param
	SessionUser                           // SESSION_USER
	DAmp                                  // &&
	AmpLt                                 // &<
	AmpGt                                 // &>
	Adjacent                              // -|-
	Xor                                   // XOR (operator)
	DStar                                 // **
	QMarkAmp                              // ?&
	QMarkPipe                             // ?|
	HashDash                              // #-
	Exclamation                           // !
	Mod                                   // %
	Placeholder                           // ?

	URIStart

	BlockStart // {%  {{
	BlockEnd   // %}  }}

	Space
	Break

	// Literals
	String
	Number
	Identifier
	Database
	Column
	ColumnDef
	Schema
	Table
	Warehouse
	Stage
	Streamlit
	Var
	BitString
	HexString
	ByteString
	NationalString
	RawString
	HeredocString
	UnicodeString

	// Data types
	Bit
	Boolean
	TinyInt
	UTinyInt
	SmallInt
	USmallInt
	MediumInt
	UMediumInt
	Int
	UInt
	BigInt
	UBigInt
	BigNum
	Int128
	UInt128
	Int256
	UInt256
	Float
	Double
	UDouble
	Decimal
	Decimal32
	Decimal64
	Decimal128
	Decimal256
	DecFloat
	UDecimal
	BigDecimal
	Char
	NChar
	VarChar
	NVarChar
	BPChar
	Text
	MediumText
	LongText
	Blob
	MediumBlob
	LongBlob
	TinyBlob
	TinyText
	Name
	Binary
	VarBinary
	JSON
	JSONB
	Time
	TimeTZ
	TimeNS
	Timestamp
	TimestampTZ
	TimestampLTZ
	TimestampNTZ
	TimestampS
	TimestampMS
	TimestampNS
	DateTime
	DateTime2
	DateTime64
	SmallDateTime
	Date
	Date32
	Int4Range
	Int4MultiRange
	Int8Range
	Int8MultiRange
	NumRange
	NumMultiRange
	TSRange
	TSMultiRange
	TSTZRange
	TSTZMultiRange
	DateRange
	DateMultiRange
	UUID
	Geography
	GeographyPoint
	Nullable
	Geometry
	Point
	Ring
	LineString
	LocalTime
	LocalTimestamp
	SysTimestamp
	MultiLineString
	Polygon
	MultiPolygon
	HLLSketch
	HStore
	Super
	Serial
	SmallSerial
	BigSerial
	XML
	Year
	UserDefined
	Money
	SmallMoney
	RowVersion
	Image
	Variant
	Object
	Inet
	IPAddress
	IPPrefix
	IPv4
	IPv6
	Enum
	Enum8
	Enum16
	FixedString
	LowCardinality
	Nested
	AggregateFunction
	SimpleAggregateFunction
	TDigest
	Unknown
	Vector
	Dynamic
	Void

	// Keywords
	Alias
	Alter
	All
	Anti
	Any
	Apply
	Array
	Asc
	Asof
	Attach
	AutoIncrement
	Begin
	Between
	BulkCollectInto
	Cache
	Case
	CharacterSet
	ClusterBy
	Collate
	Command
	Comment
	Commit
	ConnectBy
	Constraint
	Copy
	Create
	Cross
	Cube
	CurrentDate
	CurrentDatetime
	CurrentSchema
	CurrentTime
	CurrentTimestamp
	CurrentUser
	CurrentRole
	CurrentCatalog
	Declare
	Default
	Delete
	Desc
	Describe
	Detach
	Dictionary
	Distinct
	DistributeBy
	Div
	Drop
	Else
	End
	Escape
	Except
	Execute
	Exists
	False
	Fetch
	File
	FileFormat
	Filter
	Final
	First
	For
	Force
	ForeignKey
	Format
	From
	Full
	Function
	Get
	Glob
	Global
	Grant
	GroupBy
	GroupingSets
	Having
	Hint
	Ignore
	Ilike
	In
	Index
	IndexedBy
	Inner
	Insert
	Install
	Intersect
	Interval
	Into
	Introducer
	IRLike
	Is
	IsNull
	Join
	JoinMarker
	Keep
	Key
	Kill
	Language
	Lateral
	Left
	Like
	Limit
	List
	Load
	Lock
	Map
	Match
	MatchCondition
	MatchRecognize
	MemberOf
	Merge
	Model
	Natural
	Next
	Nothing
	NotNull
	Null
	ObjectIdentifier
	Offset
	On
	Only
	Operator
	OrderBy
	OrderSiblingBy
	Ordered
	Ordinality
	Out
	InOut
	Outer
	Over
	Overlaps
	Overwrite
	Partition
	PartitionBy
	Percent
	Pivot
	Positional
	Pragma
	PreWhere
	PrimaryKey
	Procedure
	Properties
	PseudoType
	Put
	Qualify
	Quote
	QDColon
	Range
	Recursive
	Refresh
	Rename
	Replace
	Returning
	Revoke
	References
	Right
	RLike
	Rollback
	Rollup
	Row
	Rows
	Select
	Semi
	Separator
	Sequence
	SerdeProperties
	Set
	Settings
	Show
	SimilarTo
	Some
	SortBy
	SoundsLike
	SQLSecurity
	StartWith
	StorageIntegration
	StraightJoin
	Struct
	Summarize
	TableSample
	Tag
	Temporary
	Top
	Then
	True
	Truncate
	Trigger
	UnCache
	Union
	Unnest
	UnPivot
	Update
	Use
	Using
	Values
	Variadic
	View
	SemanticView
	Volatile
	When
	Where
	Window
	With
	Unique
	UTCDate
	UTCTime
	UTCTimestamp
	VersionSnapshot
	TimestampSnapshot
	Option
	Sink
	Source
	Analyze
	Namespace
	Export

	// Sentinel
	HiveTokenStream
)

// AllTokenTypes returns every defined TokenType by name.
// Used in tests to detect duplicate iota values.
func AllTokenTypes() map[string]TokenType {
	return map[string]TokenType{
		"LParen": LParen, "RParen": RParen, "LBracket": LBracket,
		"RBracket": RBracket, "LBrace": LBrace, "RBrace": RBrace,
		"Comma": Comma, "Dot": Dot, "Dash": Dash, "Plus": Plus,
		"Colon": Colon, "DotColon": DotColon, "DotCaret": DotCaret,
		"DColon": DColon, "DColonDollar": DColonDollar, "DColonPercent": DColonPercent,
		"DColonQMark": DColonQMark, "DQMark": DQMark, "Semicolon": Semicolon,
		"Star": Star, "Backslash": Backslash, "Slash": Slash,
		"Lt": Lt, "Lte": Lte, "Gt": Gt, "Gte": Gte,
		"Not": Not, "Eq": Eq, "Neq": Neq, "NullsafeEq": NullsafeEq,
		"ColonEq": ColonEq, "ColonGt": ColonGt, "NColonGt": NColonGt,
		"And": And, "Or": Or, "Amp": Amp, "DPipe": DPipe, "PipeGt": PipeGt,
		"Pipe": Pipe, "PipeSlash": PipeSlash, "DPipeSlash": DPipeSlash,
		"Caret": Caret, "CaretAt": CaretAt, "Tilde": Tilde,
		"Arrow": Arrow, "DArrow": DArrow, "FArrow": FArrow,
		"Hash": Hash, "HashArrow": HashArrow, "DHashArrow": DHashArrow,
		"LRArrow": LRArrow, "DAT": DAT, "LtAt": LtAt, "AtGt": AtGt,
		"Dollar": Dollar, "Parameter": Parameter, "SessionToken": SessionToken,
		"SessionParameter": SessionParameter, "SessionUser": SessionUser,
		"DAmp": DAmp, "AmpLt": AmpLt, "AmpGt": AmpGt, "Adjacent": Adjacent,
		"Xor": Xor, "DStar": DStar, "QMarkAmp": QMarkAmp, "QMarkPipe": QMarkPipe,
		"HashDash": HashDash, "Exclamation": Exclamation, "Mod": Mod, "Placeholder": Placeholder,
		"URIStart": URIStart, "BlockStart": BlockStart, "BlockEnd": BlockEnd,
		"Space": Space, "Break": Break,
		"String": String, "Number": Number, "Identifier": Identifier,
		"Database": Database, "Column": Column, "ColumnDef": ColumnDef,
		"Schema": Schema, "Table": Table, "Warehouse": Warehouse,
		"Stage": Stage, "Streamlit": Streamlit, "Var": Var,
		"BitString": BitString, "HexString": HexString, "ByteString": ByteString,
		"NationalString": NationalString, "RawString": RawString,
		"HeredocString": HeredocString, "UnicodeString": UnicodeString,
		"Bit": Bit, "Boolean": Boolean, "TinyInt": TinyInt, "UTinyInt": UTinyInt,
		"SmallInt": SmallInt, "USmallInt": USmallInt, "MediumInt": MediumInt,
		"UMediumInt": UMediumInt, "Int": Int, "UInt": UInt, "BigInt": BigInt,
		"UBigInt": UBigInt, "BigNum": BigNum, "Int128": Int128, "UInt128": UInt128,
		"Int256": Int256, "UInt256": UInt256, "Float": Float, "Double": Double,
		"UDouble": UDouble, "Decimal": Decimal, "Decimal32": Decimal32,
		"Decimal64": Decimal64, "Decimal128": Decimal128, "Decimal256": Decimal256,
		"DecFloat": DecFloat, "UDecimal": UDecimal, "BigDecimal": BigDecimal,
		"Char": Char, "NChar": NChar, "VarChar": VarChar, "NVarChar": NVarChar,
		"BPChar": BPChar, "Text": Text, "MediumText": MediumText, "LongText": LongText,
		"Blob": Blob, "MediumBlob": MediumBlob, "LongBlob": LongBlob,
		"TinyBlob": TinyBlob, "TinyText": TinyText, "Name": Name,
		"Binary": Binary, "VarBinary": VarBinary, "JSON": JSON, "JSONB": JSONB,
		"Time": Time, "TimeTZ": TimeTZ, "TimeNS": TimeNS,
		"Timestamp": Timestamp, "TimestampTZ": TimestampTZ,
		"TimestampLTZ": TimestampLTZ, "TimestampNTZ": TimestampNTZ,
		"TimestampS": TimestampS, "TimestampMS": TimestampMS, "TimestampNS": TimestampNS,
		"DateTime": DateTime, "DateTime2": DateTime2, "DateTime64": DateTime64,
		"SmallDateTime": SmallDateTime, "Date": Date, "Date32": Date32,
		"Int4Range": Int4Range, "Int4MultiRange": Int4MultiRange,
		"Int8Range": Int8Range, "Int8MultiRange": Int8MultiRange,
		"NumRange": NumRange, "NumMultiRange": NumMultiRange,
		"TSRange": TSRange, "TSMultiRange": TSMultiRange,
		"TSTZRange": TSTZRange, "TSTZMultiRange": TSTZMultiRange,
		"DateRange": DateRange, "DateMultiRange": DateMultiRange,
		"UUID": UUID, "Geography": Geography, "GeographyPoint": GeographyPoint,
		"Nullable": Nullable, "Geometry": Geometry, "Point": Point, "Ring": Ring,
		"LineString": LineString, "LocalTime": LocalTime, "LocalTimestamp": LocalTimestamp,
		"SysTimestamp": SysTimestamp, "MultiLineString": MultiLineString,
		"Polygon": Polygon, "MultiPolygon": MultiPolygon, "HLLSketch": HLLSketch,
		"HStore": HStore, "Super": Super, "Serial": Serial, "SmallSerial": SmallSerial,
		"BigSerial": BigSerial, "XML": XML, "Year": Year, "UserDefined": UserDefined,
		"Money": Money, "SmallMoney": SmallMoney, "RowVersion": RowVersion,
		"Image": Image, "Variant": Variant, "Object": Object, "Inet": Inet,
		"IPAddress": IPAddress, "IPPrefix": IPPrefix, "IPv4": IPv4, "IPv6": IPv6,
		"Enum": Enum, "Enum8": Enum8, "Enum16": Enum16,
		"FixedString": FixedString, "LowCardinality": LowCardinality, "Nested": Nested,
		"AggregateFunction":       AggregateFunction,
		"SimpleAggregateFunction": SimpleAggregateFunction,
		"TDigest":                 TDigest, "Unknown": Unknown, "Vector": Vector,
		"Dynamic": Dynamic, "Void": Void,
		"Alias": Alias, "Alter": Alter, "All": All, "Anti": Anti, "Any": Any,
		"Apply": Apply, "Array": Array, "Asc": Asc, "Asof": Asof, "Attach": Attach,
		"AutoIncrement": AutoIncrement, "Begin": Begin, "Between": Between,
		"BulkCollectInto": BulkCollectInto, "Cache": Cache, "Case": Case,
		"CharacterSet": CharacterSet, "ClusterBy": ClusterBy, "Collate": Collate,
		"Command": Command, "Comment": Comment, "Commit": Commit,
		"ConnectBy": ConnectBy, "Constraint": Constraint, "Copy": Copy,
		"Create": Create, "Cross": Cross, "Cube": Cube,
		"CurrentDate": CurrentDate, "CurrentDatetime": CurrentDatetime,
		"CurrentSchema": CurrentSchema, "CurrentTime": CurrentTime,
		"CurrentTimestamp": CurrentTimestamp, "CurrentUser": CurrentUser,
		"CurrentRole": CurrentRole, "CurrentCatalog": CurrentCatalog,
		"Declare": Declare, "Default": Default, "Delete": Delete,
		"Desc": Desc, "Describe": Describe, "Detach": Detach,
		"Dictionary": Dictionary, "Distinct": Distinct, "DistributeBy": DistributeBy,
		"Div": Div, "Drop": Drop, "Else": Else, "End": End,
		"Escape": Escape, "Except": Except, "Execute": Execute, "Exists": Exists,
		"False": False, "Fetch": Fetch, "File": File, "FileFormat": FileFormat,
		"Filter": Filter, "Final": Final, "First": First, "For": For,
		"Force": Force, "ForeignKey": ForeignKey, "Format": Format,
		"From": From, "Full": Full, "Function": Function, "Get": Get,
		"Glob": Glob, "Global": Global, "Grant": Grant, "GroupBy": GroupBy,
		"GroupingSets": GroupingSets, "Having": Having, "Hint": Hint,
		"Ignore": Ignore, "Ilike": Ilike, "In": In, "Index": Index,
		"IndexedBy": IndexedBy, "Inner": Inner, "Insert": Insert, "Install": Install,
		"Intersect": Intersect, "Interval": Interval, "Into": Into,
		"Introducer": Introducer, "IRLike": IRLike, "Is": Is, "IsNull": IsNull,
		"Join": Join, "JoinMarker": JoinMarker, "Keep": Keep, "Key": Key,
		"Kill": Kill, "Language": Language, "Lateral": Lateral, "Left": Left,
		"Like": Like, "Limit": Limit, "List": List, "Load": Load, "Lock": Lock,
		"Map": Map, "Match": Match, "MatchCondition": MatchCondition,
		"MatchRecognize": MatchRecognize, "MemberOf": MemberOf, "Merge": Merge,
		"Model": Model, "Natural": Natural, "Next": Next, "Nothing": Nothing,
		"NotNull": NotNull, "Null": Null, "ObjectIdentifier": ObjectIdentifier,
		"Offset": Offset, "On": On, "Only": Only, "Operator": Operator,
		"OrderBy": OrderBy, "OrderSiblingBy": OrderSiblingBy, "Ordered": Ordered,
		"Ordinality": Ordinality, "Out": Out, "InOut": InOut, "Outer": Outer,
		"Over": Over, "Overlaps": Overlaps, "Overwrite": Overwrite,
		"Partition": Partition, "PartitionBy": PartitionBy, "Percent": Percent,
		"Pivot": Pivot, "Positional": Positional, "Pragma": Pragma,
		"PreWhere": PreWhere, "PrimaryKey": PrimaryKey, "Procedure": Procedure,
		"Properties": Properties, "PseudoType": PseudoType, "Put": Put,
		"Qualify": Qualify, "Quote": Quote, "QDColon": QDColon,
		"Range": Range, "Recursive": Recursive, "Refresh": Refresh,
		"Rename": Rename, "Replace": Replace, "Returning": Returning,
		"Revoke": Revoke, "References": References, "Right": Right,
		"RLike": RLike, "Rollback": Rollback, "Rollup": Rollup,
		"Row": Row, "Rows": Rows, "Select": Select, "Semi": Semi,
		"Separator": Separator, "Sequence": Sequence,
		"SerdeProperties": SerdeProperties,
		"Set":             Set, "Settings": Settings, "Show": Show, "SimilarTo": SimilarTo,
		"Some": Some, "SortBy": SortBy, "SoundsLike": SoundsLike,
		"SQLSecurity": SQLSecurity, "StartWith": StartWith,
		"StorageIntegration": StorageIntegration, "StraightJoin": StraightJoin,
		"Struct": Struct, "Summarize": Summarize, "TableSample": TableSample,
		"Tag": Tag, "Temporary": Temporary, "Top": Top, "Then": Then,
		"True": True, "Truncate": Truncate, "Trigger": Trigger,
		"UnCache": UnCache, "Union": Union, "Unnest": Unnest, "UnPivot": UnPivot,
		"Update": Update, "Use": Use, "Using": Using, "Values": Values,
		"Variadic": Variadic, "View": View, "SemanticView": SemanticView,
		"Volatile": Volatile, "When": When, "Where": Where, "Window": Window,
		"With": With, "Unique": Unique, "UTCDate": UTCDate, "UTCTime": UTCTime,
		"UTCTimestamp": UTCTimestamp, "VersionSnapshot": VersionSnapshot,
		"TimestampSnapshot": TimestampSnapshot, "Option": Option,
		"Sink": Sink, "Source": Source, "Analyze": Analyze,
		"Namespace": Namespace, "Export": Export,
		"HiveTokenStream": HiveTokenStream,
	}
}
