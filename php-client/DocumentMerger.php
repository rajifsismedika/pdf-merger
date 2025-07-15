<?php

/**
 * DocumentMerger PHP Client
 * 
 * A PHP client class for interacting with the PDF Merge API service.
 * Provides methods to merge PDFs from URLs and generate reports.
 */
class DocumentMerger
{
    private $serviceUrl;
    private $timeout;
    private $headers;

    /**
     * Constructor
     * 
     * @param string $serviceUrl Base URL of the PDF merge service (default: http://localhost:8080)
     * @param int $timeout Request timeout in seconds (default: 60)
     */
    public function __construct($serviceUrl = "http://localhost:8585", $timeout = 60)
    {
        $this->serviceUrl = rtrim($serviceUrl, '/');
        $this->timeout = $timeout;
        $this->headers = array(
            'Content-Type: application/json',
            'Accept: application/pdf, application/json'
        );
    }

    /**
     * Merge multiple PDFs from URLs
     * 
     * @param array $urls Array of PDF URLs to merge
     * @param string|null $name Optional name for the merged PDF
     * @param bool $stream Whether to stream the response directly (default: false)
     * @return array Returns array with 'success' boolean, 'data' (PDF content or error), and 'filename'
     */
    public function mergePDFs($urls, $name = null, $stream = false)
    {
        if (empty($urls)) {
            return array(
                'success' => false,
                'error' => 'URLs array cannot be empty',
                'data' => null
            );
        }

        $payload = array(
            'urls' => $urls
        );

        if ($name !== null) {
            $payload['name'] = $name;
        }

        return $this->makeRequest('POST', '/merge', $payload, $stream);
    }

    /**
     * Generate and download a report by ID
     * 
     * @param string $reportId The report ID
     * @param bool $stream Whether to stream the response directly
     * @return array Returns array with 'success' boolean, 'data' (PDF content or error), and 'filename'
     */
    public function generateReport($reportId, $stream = false)
    {
        if (empty($reportId)) {
            return array(
                'success' => false,
                'error' => 'Report ID cannot be empty',
                'data' => null
            );
        }

        return $this->makeRequest('GET', "/report/{$reportId}", null, $stream);
    }

    /**
     * Save PDF data to file
     * 
     * @param string $pdfData Binary PDF data
     * @param string $filepath Path where to save the file
     * @return bool Returns true on success, false on failure
     */
    public function savePDF($pdfData, $filepath)
    {
        try {
            $directory = dirname($filepath);
            if (!is_dir($directory)) {
                mkdir($directory, 0755, true);
            }

            return file_put_contents($filepath, $pdfData) !== false;
        } catch (Exception $e) {
            return false;
        }
    }

    /**
     * Download PDF and save to file
     * 
     * @param array $urls Array of PDF URLs to merge
     * @param string $filepath Path where to save the merged PDF
     * @param string|null $name Optional name for the merged PDF
     * @return array Returns result array with success status
     */
    public function downloadMergedPDF($urls, $filepath, $name = null)
    {
        $result = $this->mergePDFs($urls, $name, false);
        
        if ($result['success']) {
            if ($this->savePDF($result['data'], $filepath)) {
                return array(
                    'success' => true,
                    'filepath' => $filepath,
                    'filename' => $result['filename'],
                    'size' => $result['size']
                );
            } else {
                return array(
                    'success' => false,
                    'error' => 'Failed to save PDF to file'
                );
            }
        }

        return $result;
    }

    /**
     * Download report and save to file
     * 
     * @param string $reportId The report ID
     * @param string $filepath Path where to save the report PDF
     * @return array Returns result array with success status
     */
    public function downloadReport($reportId, $filepath)
    {
        $result = $this->generateReport($reportId, false);
        
        if ($result['success']) {
            if ($this->savePDF($result['data'], $filepath)) {
                return array(
                    'success' => true,
                    'filepath' => $filepath,
                    'filename' => $result['filename'],
                    'size' => $result['size']
                );
            } else {
                return array(
                    'success' => false,
                    'error' => 'Failed to save PDF to file'
                );
            }
        }

        return $result;
    }

    /**
     * Set request timeout
     * 
     * @param int $timeout Timeout in seconds
     */
    public function setTimeout($timeout)
    {
        $this->timeout = $timeout;
    }

    /**
     * Add custom header
     * 
     * @param string $header Header string (e.g., "Authorization: Bearer token")
     */
    public function addHeader($header)
    {
        $this->headers[] = $header;
    }

    /**
     * Make HTTP request to the service
     * 
     * @param string $method HTTP method (GET, POST)
     * @param string $endpoint API endpoint
     * @param array|null $data Request data for POST requests
     * @param bool $stream Whether to stream the response directly
     * @return array Returns response array
     */
    private function makeRequest($method, $endpoint, $data = null, $stream = false)
    {
        $url = $this->serviceUrl . $endpoint;
        $ch = curl_init();

        // Set up basic curl options
        curl_setopt_array($ch, array(
            CURLOPT_URL => $url,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT => $this->timeout,
            CURLOPT_FOLLOWLOCATION => true,
            CURLOPT_HEADER => true,
            CURLOPT_CUSTOMREQUEST => $method,
            CURLOPT_HTTPHEADER => $this->headers,
        ));

        // Handle POST data
        if ($method === 'POST' && $data !== null) {
            $jsonData = json_encode($data);
            curl_setopt($ch, CURLOPT_POSTFIELDS, $jsonData);
            curl_setopt($ch, CURLOPT_POST, true);
        }

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
        $error = curl_error($ch);
        curl_close($ch);

        if ($response === false) {
            return array(
                'success' => false,
                'error' => 'cURL error: ' . $error,
                'data' => null
            );
        }

        $headers = substr($response, 0, $headerSize);
        $body = $response;

        if ($httpCode >= 200 && $httpCode < 300) {
            // If streaming is requested, output PDF directly to browser
            if ($stream) {
                $filename = $this->extractFilename($headers) ?: ($data['name'] ?: 'merged.pdf');
                header('Content-Type: application/pdf');
                header('Content-Disposition: inline; filename="' . $filename . '"');
                header('Content-Length: ' . strlen($body));
                echo $body;
                exit;
            }

            // Return the PDF content for further processing
            return array(
                'success' => true,
                'data' => $body,
                'headers' => $headers,
                'filename' => $this->extractFilename($headers),
                'size' => strlen($body)
            );
        }

        // Handle error response
        $errorData = json_decode($body, true);
        $errorMessage = isset($errorData['error']) ? $errorData['error'] : $body;

        return array(
            'success' => false,
            'error' => "HTTP {$httpCode}: {$errorMessage}",
            'data' => null,
            'http_code' => $httpCode
        );
    }

    /**
     * Extract filename from Content-Disposition header
     * 
     * @param string $headers Response headers
     * @return string|null Extracted filename or null
     */
    private function extractFilename($headers)
    {
        if (preg_match('/filename[^;=\n]*=(([\'"]).*?\2|[^;\n]*)/', $headers, $matches)) {
            return trim($matches[1], '"\'');
        }
        return null;
    }
}
