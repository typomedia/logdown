<?php

namespace AppBundle\Repository;

use Symfony\Component\Filesystem\Exception\IOException;

class LogRepository
{
    /**
     * @var string $path
     */
    protected $path;

    /**
     * @param $path
     */
    public function __construct($path)
    {
        $this->path = $path . '/var/scripts/';
    }

    /**
     * @param $name
     * @return false|string
     */
    public function getContent($name) {
        return file_get_contents($this->path . $name);

    }
}