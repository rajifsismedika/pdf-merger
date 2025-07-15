<?php

require_once 'DocumentMerger.php';

// Initialize DocumentMerger with correct service URL
$merger = new DocumentMerger("http://host.docker.internal:8585");

// Check if we should stream directly to browser or return data
$stream = isset($_REQUEST['stream']) && $_REQUEST['stream'] === 'true';

// Merge PDFs
$result = $merger->mergePDFs($urls, $filename, $stream);

// If streaming was requested, the response was already sent
if ($stream) {
    return;
}

// Handle the result
if ($result['success']) {
    // Set headers for PDF download
    header('Content-Type: application/pdf');
    header('Content-Disposition: inline; filename="' . ($result['filename'] ?: $filename . '.pdf') . '"');
    header('Content-Length: ' . $result['size']);
    
    // Output PDF content
    echo $result['data'];
} else {
    // Return error as JSON
    http_response_code(500);
    header('Content-Type: application/json');
    echo json_encode(array(
        'success' => false,
        'error' => $result['error']
    ));
}
