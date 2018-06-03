<?php
/**
 * High-performance PHP process supervisor and load balancer written in Go
 *
 * @author Wolfy-J
 */


/**
 * Class PSR7Worker serves PSR-7 requests and consume responses.
 */
class PSR7Worker
{
    /**
     * @var \Spiral\RoadRunner\Worker
     */
    private $worker;

    /**
     * @param \Spiral\RoadRunner\Worker $worker
     */
    public function __construct(\Spiral\RoadRunner\Worker $worker)
    {
        $this->worker = $worker;
    }

    /**
     * @return \Spiral\RoadRunner\Worker
     */
    public function getWorker(): \Spiral\RoadRunner\Worker
    {
        return $this->worker;
    }

    /**
     * @return \Psr\Http\Message\ServerRequestInterface|null
     */
    public function acceptRequest()
    {
        $body = $this->worker->receive($ctx);
        if (empty($body) && empty($ctx)) {
            // termination request
            return null;
        }

        parse_str($ctx['rawQuery'], $query);

        $body = 'php://input';
        $parsedBody = null;
        if ($ctx['parsed']) {
            $parsedBody = json_decode($body, true);
        } elseif ($body != null) {
            $parsedBody = new \Zend\Diactoros\Stream("php://memory", "rwb");
            $parsedBody->write($body);
        }

        return new \Zend\Diactoros\ServerRequest(
            $_SERVER,
            $this->wrapUploads($ctx['uploads']),
            $ctx['uri'],
            $ctx['method'],
            $body,
            $ctx['headers'],
            $ctx['cookies'],
            $query,
            $parsedBody,
            $ctx['protocol']
        );
    }

    /**
     * Send response to the application server.
     *
     * @param \Psr\Http\Message\ResponseInterface $response
     */
    public function respond(\Psr\Http\Message\ResponseInterface $response)
    {
        $this->worker->error("asd");

        $this->worker->send($response->getBody(), json_encode([
            'status'  => $response->getStatusCode(),
            'headers' => $response->getHeaders()
        ]));
    }

    /**
     * Wraps all uploaded files with UploadedFile.
     *
     * @param array $files
     *
     * @return array
     */
    private function wrapUploads($files): array
    {
        if (empty($files)) {
            return [];
        }

        $result = [];
        foreach ($files as $index => $file) {
            if (isset($file['name'])) {
                $result[$index] = new \Zend\Diactoros\UploadedFile(
                    $file['tmpName'],
                    $file['size'],
                    $file['error'],
                    $file['name'],
                    $file['type']
                );
                continue;
            }

            $result[$index] = $this->wrapUploads($file);
        }

        return $result;
    }
}