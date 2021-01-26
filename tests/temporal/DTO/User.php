<?php


namespace Temporal\Tests\DTO;

use Temporal\Internal\Marshaller\Meta\Marshal;

class User
{
    #[Marshal(name: "Name")]
    public string $name;

    #[Marshal(name: "Email")]
    public string $email;
}