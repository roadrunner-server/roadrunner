<?php

namespace Spiral\RoadRunner;

use Psr\Log\InvalidArgumentException;
use Psr\Log\LoggerInterface;
use Psr\Log\LogLevel;
use Spiral\Goridge\RPC;

final class RPCLogger implements LoggerInterface
{
    private const PSR_TO_LOGRUS_LEVELS = [
        // levels above "error" causes logrus to os.exit() or panic()
        LogLevel::EMERGENCY => "error",
        LogLevel::ALERT => "error",
        LogLevel::CRITICAL => "error",
        LogLevel::ERROR => "error",
        LogLevel::WARNING => "warning",
        LogLevel::NOTICE => "info",
        LogLevel::INFO => "info",
        LogLevel::DEBUG => "debug",
    ];

    /**
     * @var RPC
     */
    private $rpc;

    public function __construct(RPC $rpc)
    {
        $this->rpc = $rpc;
    }

    /**
     * @inheritDoc
     */
    public function emergency($message, array $context = array())
    {
        $this->log(LogLevel::EMERGENCY, $message, $context);
    }

    /**
     * @inheritDoc
     */
    public function alert($message, array $context = array())
    {
        $this->log(LogLevel::ALERT, $message, $context);
    }

    /**
     * @inheritDoc
     */
    public function critical($message, array $context = array())
    {
        $this->log(LogLevel::CRITICAL, $message, $context);
    }

    /**
     * @inheritDoc
     */
    public function error($message, array $context = array())
    {
        $this->log(LogLevel::ERROR, $message, $context);
    }

    /**
     * @inheritDoc
     */
    public function warning($message, array $context = array())
    {
        $this->log(LogLevel::WARNING, $message, $context);
    }

    /**
     * @inheritDoc
     */
    public function notice($message, array $context = array())
    {
        $this->log(LogLevel::NOTICE, $message, $context);
    }

    /**
     * @inheritDoc
     */
    public function info($message, array $context = array())
    {
        $this->log(LogLevel::INFO, $message, $context);
    }

    /**
     * @inheritDoc
     */
    public function debug($message, array $context = array())
    {
        $this->log(LogLevel::DEBUG, $message, $context);
    }

    /**
     * @inheritDoc
     */
    public function log($level, $message, array $context = array())
    {
        if (!isset(self::PSR_TO_LOGRUS_LEVELS[$level])) {
            throw new InvalidArgumentException(sprintf("Invalid level '%s': only constants of '%s' are allowed", $level, LogLevel::class));
        }

        $context['psr_level'] = $level;

        $this->rpc->call("log.Log", [
            "message" => $message,
            "level" => self::PSR_TO_LOGRUS_LEVELS[$level],
            "fields" => $context,
        ]);
    }
}