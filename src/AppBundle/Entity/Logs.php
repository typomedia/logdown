<?php

namespace AppBundle\Entity;

use Doctrine\ORM\Mapping as ORM;

/**
 * Logs
 *
 * @ORM\Table(name="Logs")
 * @ORM\Entity
 */
class Logs
{
    /**
     * @var int|null
     *
     * @ORM\Column(name="Id", type="integer", nullable=true)
     * @ORM\Id
     * @ORM\GeneratedValue(strategy="IDENTITY")
     */
    private $id;

    /**
     * @var \DateTime|null
     *
     * @ORM\Column(name="Date", type="date", nullable=true)
     */
    private $date;

    /**
     * @var \DateTime|null
     *
     * @ORM\Column(name="Time", type="time", nullable=true)
     */
    private $time;

    /**
     * @var string|null
     *
     * @ORM\Column(name="Server", type="text", nullable=true)
     */
    private $server;

    /**
     * @var string|null
     *
     * @ORM\Column(name="Method", type="text", nullable=true)
     */
    private $method;

    /**
     * @var string|null
     *
     * @ORM\Column(name="Request", type="text", nullable=true)
     */
    private $request;

    /**
     * @var string|null
     *
     * @ORM\Column(name="Param", type="text", nullable=true)
     */
    private $param;

    /**
     * @var int|null
     *
     * @ORM\Column(name="Port", type="integer", nullable=true)
     */
    private $port;

    /**
     * @var string|null
     *
     * @ORM\Column(name="User", type="text", nullable=true)
     */
    private $user;

    /**
     * @var string|null
     *
     * @ORM\Column(name="Client", type="text", nullable=true)
     */
    private $client;

    /**
     * @var string|null
     *
     * @ORM\Column(name="Agent", type="text", nullable=true)
     */
    private $agent;

    /**
     * @var string|null
     *
     * @ORM\Column(name="Referer", type="text", nullable=true)
     */
    private $referer;

    /**
     * @var int|null
     *
     * @ORM\Column(name="Status", type="integer", nullable=true)
     */
    private $status;

    /**
     * @var int|null
     *
     * @ORM\Column(name="Substatus", type="integer", nullable=true)
     */
    private $substatus;

    /**
     * @var string|null
     *
     * @ORM\Column(name="Win32", type="text", nullable=true)
     */
    private $win32;

    /**
     * @var int|null
     *
     * @ORM\Column(name="Duration", type="integer", nullable=true)
     */
    private $duration;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getDate(): ?\DateTimeInterface
    {
        return $this->date;
    }

    public function setDate(?\DateTimeInterface $date): self
    {
        $this->date = $date;

        return $this;
    }

    public function getTime(): ?\DateTimeInterface
    {
        return $this->time;
    }

    public function setTime(?\DateTimeInterface $time): self
    {
        $this->time = $time;

        return $this;
    }

    public function getServer(): ?string
    {
        return $this->server;
    }

    public function setServer(?string $server): self
    {
        $this->server = $server;

        return $this;
    }

    public function getMethod(): ?string
    {
        return $this->method;
    }

    public function setMethod(?string $method): self
    {
        $this->method = $method;

        return $this;
    }

    public function getRequest(): ?string
    {
        return $this->request;
    }

    public function setRequest(?string $request): self
    {
        $this->request = $request;

        return $this;
    }

    public function getParam(): ?string
    {
        return $this->param;
    }

    public function setParam(?string $param): self
    {
        $this->param = $param;

        return $this;
    }

    public function getPort(): ?int
    {
        return $this->port;
    }

    public function setPort(?int $port): self
    {
        $this->port = $port;

        return $this;
    }

    public function getUser(): ?string
    {
        return $this->user;
    }

    public function setUser(?string $user): self
    {
        $this->user = $user;

        return $this;
    }

    public function getClient(): ?string
    {
        return $this->client;
    }

    public function setClient(?string $client): self
    {
        $this->client = $client;

        return $this;
    }

    public function getAgent(): ?string
    {
        return $this->agent;
    }

    public function setAgent(?string $agent): self
    {
        $this->agent = $agent;

        return $this;
    }

    public function getReferer(): ?string
    {
        return $this->referer;
    }

    public function setReferer(?string $referer): self
    {
        $this->referer = $referer;

        return $this;
    }

    public function getStatus(): ?int
    {
        return $this->status;
    }

    public function setStatus(?int $status): self
    {
        $this->status = $status;

        return $this;
    }

    public function getSubstatus(): ?int
    {
        return $this->substatus;
    }

    public function setSubstatus(?int $substatus): self
    {
        $this->substatus = $substatus;

        return $this;
    }

    public function getWin32(): ?string
    {
        return $this->win32;
    }

    public function setWin32(?string $win32): self
    {
        $this->win32 = $win32;

        return $this;
    }

    public function getDuration(): ?int
    {
        return $this->duration;
    }

    public function setDuration(?int $duration): self
    {
        $this->duration = $duration;

        return $this;
    }


}
