package files

func openAINativeDelivery(kind FileKind) (DeliveryKind, bool) {
	switch kind {
	case KindPDF:
		return DeliveryPDF, true
	case KindImage:
		return DeliveryImage, true
	case KindDocx, KindOffice, KindSpreadsheet,
		KindText, KindCode, KindMarkdown,
		KindJSON, KindHTML, KindXML:
		return DeliveryProviderFile, true
	default:
		return "", false
	}
}

func anthropicNativeDelivery(kind FileKind) (DeliveryKind, bool) {
	switch kind {
	case KindPDF:
		return DeliveryPDF, true
	case KindImage:
		return DeliveryImage, true
	default:
		return "", false
	}
}

func anthropicNeedsInline(kind FileKind) bool {
	switch kind {
	case KindPDF, KindImage:
		return false
	case KindExecutable, KindMedia, KindArchiveUnsupported,
		KindLegacyDoc, KindUnknown:
		return false
	default:
		return true
	}
}
